package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type TestCase struct {
	Description string `json:"description"`
	Calls       []Call `json:"calls"`
}

type Call struct {
	On       On                `json:"on"`
	Expect   Expect            `json:"expect"`
	Remember map[string]string `json:"remember"`
}

type On struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

type Expect struct {
	StatusCode  int                    `json:"statusCode"`
	ContentType string                 `json:"contentType"`
	Body        map[string]interface{} `json:"body"`
	BodySchema  string                 `json:"bodySchema"`
}

var (
	suiteDir = flag.String("d", ".", "Path to the directory that contains test suite.")
	host     = flag.String("h", "http://localhost:8080", "Test server address")
)

func main() {
	flag.Parse()

	loader := testCaseLoader{}
	testCases, err := loader.loadDir(*suiteDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test Cases: %v\n", testCases)

	call(testCases[0], testCases[0].Calls[0]) //TODO cycle
}

type testCaseLoader struct {
	tests []TestCase
}

func (s *testCaseLoader) loadDir(dir string) ([]TestCase, error) {
	err := filepath.Walk(dir, s.loadFile)
	if err != nil {
		return nil, err
	}

	return s.tests, nil
}

func (s *testCaseLoader) loadFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	if info.IsDir() {
		return nil
	}

	if !strings.HasSuffix(info.Name(), ".json") {
		return nil
	}

	fmt.Printf("Process file: %s\n", info.Name())
	content, e := ioutil.ReadFile(path)

	if e != nil {
		fmt.Printf("File error: %v\n", e)
		return nil
	}

	var testCases []TestCase
	err = json.Unmarshal(content, &testCases)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return nil
	}

	s.tests = append(s.tests, testCases...)
	return nil
}

func call(testCase TestCase, call Call) (rememberMap map[string]string, failedExpectations []string) {
	on := call.On

	req, _ := http.NewRequest(on.Method, *host+on.Url, nil) //TODO extract url to param

	for key, value := range on.Headers {
		req.Header.Add(key, value)
	}

	q := req.URL.Query()
	for key, value := range on.Params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error when sending request")
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response")
		return
	}

	//fmt.Printf("Code: %v\n", resp.Status)
	fmt.Printf("Resp: %v\n", string(body))

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		fmt.Println("Error parsing body")
		return
	}

	testResp := Response{http: *resp, body: string(body)}
	reporter := NewConsoleReporter()
	result := TestResult{Case: testCase, Resp: testResp}

	exps := expectations(call)
	for _, exp := range exps {
		checkErr := exp.check(testResp)
		if checkErr != nil {
			result.Cause = checkErr
			break
		}
	}
	reporter.Report(result)
	//fmt.Printf("bm: %v\n", bodyMap)

	// v := getByPath(bodyMap, "token")
	//fmt.Printf("v: %v\n", v)

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		fmt.Println("Error parsing body")
		return
	}

	rememberMap, err = remember(bodyMap, call.Remember)
	fmt.Printf("rememberMap: %v\n", rememberMap)
	if err != nil {
		fmt.Println("Error remember")
		return
	}

	return
}

func expectations(call Call) []ResponseExpectation {
	var exps []ResponseExpectation

	if call.Expect.StatusCode != -1 {
		exps = append(exps, StatusExpectation{statusCode: call.Expect.StatusCode})
	}

	if call.Expect.BodySchema != "" {
		// for now use path relative to suiteDir
		uri, err := filepath.Abs(*suiteDir)
		if err != nil {
			fmt.Println(err)
		}

		uri = "file:///" + filepath.ToSlash(filepath.Join(uri, call.Expect.BodySchema))
		exps = append(exps, BodySchemaExpectation{schemaURI: uri})
	}
	// and so on
	return exps
}

func remember(bodyMap map[string]interface{}, remember map[string]string) (rememberedMap map[string]string, err error) {
	rememberedMap = make(map[string]string)

	for varName, path := range remember {
		if strings.HasPrefix(path, "body.") {
			bodyPath := strings.TrimPrefix(path, "body.")
			splitPath := strings.Split(bodyPath, ".")

			b := make([]interface{}, len(splitPath))
			for i := range splitPath {
				b[i] = splitPath[i]
			}

			rememberVar := getByPath(bodyMap, b...)
			if rememberVar != nil {
				rememberedMap[varName] = rememberVar.(string)
			} else {
				err = errors.New("Remembered value not found: %v\n")
			}
			//fmt.Printf("v: %v\n", getByPath(bodyMap, b...))
		}
	}

	return rememberedMap, err
}

func getByPath(m interface{}, path ...interface{}) interface{} {

	for _, p := range path {
		switch idx := p.(type) {
		case string:
			m = m.(map[string]interface{})[idx]
		case int:
			m = m.([]interface{})[idx]
		}
	}
	return m
}

type TestResult struct {
	Case TestCase
	Resp Response
	// in case test failed, cause must be specified
	Cause error
}

type Response struct {
	http http.Response
	body string
}

type ResponseExpectation interface {
	check(resp Response) error
}

type StatusExpectation struct {
	statusCode int
}

func (e StatusExpectation) check(resp Response) error {
	if resp.http.StatusCode != e.statusCode {
		msg := fmt.Sprintf("Unexpected Status Code. Expected: %d, Actual: %d", e.statusCode, resp.http.StatusCode)
		return errors.New(msg)
	}
	return nil
}

type BodySchemaExpectation struct {
	schemaURI string
}

func (e BodySchemaExpectation) check(resp Response) error {
	schemaLoader := gojsonschema.NewReferenceLoader(e.schemaURI)
	documentLoader := gojsonschema.NewStringLoader(resp.body)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if !result.Valid() {
		msg := "Unexpected Body Schema:\n"
		for _, desc := range result.Errors() {
			msg = fmt.Sprintf(msg+"\n\t%s\n", desc)
		}
		return errors.New(msg)
	}

	return nil
}

// TODO remember
// TODO resp matches schema
// TODO expect matchers: equal, anyOf, arrHasSize, arrHasItems
// TODO expect matchers without <any> indexes
// TODO expect: header, statusCode, body
// TODO add company name to test case (track snapshot usage)
