package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clbanning/mxj"
)

var (
	suiteDir = flag.String("d", ".", "Path to the directory that contains test suite.")
	host     = flag.String("h", "http://localhost:8080", "Test server address")
	verbose  = flag.Bool("v", false, "Verbose mode")
)

func main() {
	flag.Parse()

	loader := testCaseLoader{}
	suits, err := loader.loadDir(*suiteDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	path, _ := filepath.Abs("./report")
	reporter := NewMultiReporter(NewJUnitReporter(path), NewConsoleReporter())

	// test case runner?
	for _, suite := range suits {
		for _, testCase := range suite.Cases {

			rememberedMap := make(map[string]string)

			for _, c := range testCase.Calls {
				tr := call(suite, testCase, c, rememberedMap)
				tr.Suite = suite
				reporter.Report(*tr)
			}
		}
	}

	reporter.Flush()
}

func call(testSuite TestSuite, testCase TestCase, call Call, rememberMap map[string]string) (result *TestResult) {
	debugMsg("--- Starting call ...") // TODO add call description
	start := time.Now()
	result = &TestResult{Case: testCase}

	on := call.On

	dat := []byte(on.Body)
	if on.BodyFile != "" {
		uri, err := getTestAssetUri(testSuite.Dir, on.BodyFile)
		if err != nil {
			result.Cause = err
			return
		}

		if d, err := ioutil.ReadFile(uri); err == nil {
			dat = d
		} else {
			result.Cause = fmt.Errorf("Can't read body file: %s", err.Error())
			return
		}
	}

	req := populateRequest(on, string(dat), rememberMap)
	debugMsg("Request: ", req)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		debugMsg("Error when sending request", err)
		result.Cause = err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		debugMsg("Error reading response")
		result.Cause = err
		return
	}

	debugMsg("Resp: ", string(body))
	end := time.Now()

	testResp := Response{http: *resp, body: body}
	result.Resp = testResp
	result.Duration = end.Sub(start)

	exps, err := expectations(call, testSuite.Dir)
	if err != nil {
		result.Cause = err
		return
	}

	for _, exp := range exps {
		checkErr := exp.check(testResp)
		if checkErr != nil {
			result.Cause = checkErr
			return
		}
	}

	m, err := testResp.bodyAsMap()
	if err != nil {
		debugMsg("Can't parse response body to Map for [remember]")
		result.Cause = err
		return
	}

	err = remember(m, call.Remember, rememberMap)
	debugMsg("Remember: ", rememberMap)
	if err != nil {
		debugMsg("Error remember")
		result.Cause = err
		return
	}

	return result
}

func populateRequest(on On, body string, rememberMap map[string]string) *http.Request {

	url := urlPrefix(on.URL)

	body = populateRememberedVars(body, rememberMap)
	dat := []byte(body)

	req, _ := http.NewRequest(on.Method, url, bytes.NewBuffer(dat))

	for key, value := range on.Headers {
		req.Header.Add(key, populateRememberedVars(value, rememberMap))
	}

	q := req.URL.Query()
	for key, value := range on.Params {
		q.Add(key, populateRememberedVars(value, rememberMap))
	}
	req.URL.RawQuery = q.Encode()

	return req
}

func urlPrefix(url string) string {
	res := *host + url
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		url = url
	}

	return res
}

func populateRememberedVars(str string, rememberMap map[string]string) string {
	res := str
	for varName, val := range rememberMap {
		placeholder := "{" + varName + "}"
		res = strings.Replace(res, placeholder, val, -1)
	}
	return res
}

func expectations(call Call, srcDir string) ([]ResponseExpectation, error) {
	var exps []ResponseExpectation
	if call.Expect.StatusCode != 0 {
		exps = append(exps, StatusExpectation{statusCode: call.Expect.StatusCode})
	}

	if call.Expect.BodySchema != "" {
		uri, err := getTestAssetUri(srcDir, call.Expect.BodySchema)
		if err != nil {
			return nil, err
		}
		exps = append(exps, BodySchemaExpectation{schemaURI: uri})
	}

	if len(call.Expect.Body) > 0 {
		exps = append(exps, BodyExpectation{pathExpectations: call.Expect.Body})
	}

	if len(call.Expect.Headers) > 0 {
		for k, v := range call.Expect.Headers {
			exps = append(exps, HeaderExpectation{Name: k, Value: v})
		}
	}

	if call.Expect.ContentType != "" {
		parser := func(value string) string {
			contentType, _, _ := mime.ParseMediaType(value)
			return contentType
		}
		exps = append(exps, HeaderExpectation{
			Name:        "content-type",
			Value:       call.Expect.ContentType,
			ValueParser: parser,
		})
	}

	// and so on
	return exps, nil
}

func getTestAssetUri(srcDir string, assetPath string) (string, error) {
	if filepath.IsAbs(assetPath) {
		// ignore srcDir
		return assetPath, nil
	}

	uri, err := filepath.Abs(filepath.Join(*suiteDir, srcDir, assetPath))
	if err != nil {
		return "", errors.New("Invalid file path: " + assetPath)
	}

	return filepath.ToSlash(uri), nil
}

func remember(bodyMap map[string]interface{}, remember map[string]string, rememberedMap map[string]string) (err error) {

	for varName, path := range remember {

		splitPath := strings.Split(path, ".")

		if rememberVar, err := getByPath(bodyMap, splitPath...); err == nil {
			rememberedMap[varName] = rememberVar
		} else {
			strErr := fmt.Sprintf("Remembered value not found, path: %v", path)
			err = errors.New(strErr)
		}
		//fmt.Printf("v: %v\n", getByPath(bodyMap, b...))
	}

	return err
}

func (e Response) bodyAsMap() (map[string]interface{}, error) {
	var bodyMap map[string]interface{}
	var err error

	contentType, _, _ := mime.ParseMediaType(e.http.Header.Get("content-type"))
	if contentType == "application/xml" || contentType == "text/xml" {
		m, err := mxj.NewMapXml(e.body)
		if err == nil {
			bodyMap = m.Old() // cast to map
		}
	}

	if contentType == "application/json" {
		err = json.Unmarshal(e.body, &bodyMap)
	}

	return bodyMap, err
}

func debugMsg(a ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Print("\t")
	fmt.Println(a...)
}

func debugMsgF(tpl string, a ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Printf(tpl, a...)
}

// TODO finilize suite schema

// optional/under discussion
// TODO matchers: not() ?
// TODO rename remember > keep or memo ?
// TODO full body expectation from file (security testing)
// TODO concurrent tests run
