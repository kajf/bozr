package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/xeipuuv/gojsonschema"
)

// TestSuite represents file with test cases.
type TestSuite struct {
	// file name
	Name string
	// path to a file
	PackageName string
	// test cases listed in a file
	Cases []TestCase
}

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
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Params   map[string]string `json:"params"`
	Body     string            `json:"body"`
	BodyFile string            `json:"bodyFile"`
}

type Expect struct {
	StatusCode  int               `json:"statusCode"`
	ContentType string            `json:"contentType"`
	Body        map[string]string `json:"body"`
	BodySchema  string            `json:"bodySchema"`
}

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

	rememberedMap := make(map[string]string)

	path, _ := filepath.Abs("./report")
	reporter := NewMultiReporter(NewJUnitReporter(path), NewConsoleReporter())

	// test case runner?
	for _, suite := range suits {
		for _, testCase := range suite.Cases {
			for _, c := range testCase.Calls {
				tr, err := call(testCase, c, rememberedMap)
				if err != nil {
					panic(err)
				}
				tr.Suite = suite
				reporter.Report(*tr)
			}
		}
	}

	reporter.Flush()
}

type testCaseLoader struct {
	suits []TestSuite
}

func (s *testCaseLoader) loadDir(dir string) ([]TestSuite, error) {
	err := filepath.Walk(dir, s.loadFile)
	if err != nil {
		return nil, err
	}

	return s.suits, nil
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

	absPath, err := filepath.Abs(*suiteDir)
	if err != nil {
		return nil
	}

	pack := strings.TrimSuffix(strings.TrimPrefix(path, absPath), info.Name())
	name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
	su := TestSuite{Name: name, PackageName: pack, Cases: testCases}
	s.suits = append(s.suits, su)
	return nil
}

func call(testCase TestCase, call Call, rememberMap map[string]string) (*TestResult, error) {
	debugMsg("--- Starting call ...") // TODO add call description
	start := time.Now()

	on := call.On

	dat := []byte(on.Body)
	if on.BodyFile != "" {
		uri := getFileUti(*suiteDir, on.BodyFile)
		if d, err := ioutil.ReadFile(uri); err == nil{
			dat = d
		} else {
			debugMsg("Can't read body file: ", err.Error())
		}
	}

	req, _ := http.NewRequest(on.Method, *host + on.URL, bytes.NewBuffer(dat))

	for key, value := range on.Headers {
		req.Header.Add(key, putRememberedVars(value, rememberMap))
	}

	q := req.URL.Query()
	for key, value := range on.Params {
		q.Add(key, putRememberedVars(value, rememberMap))
	}
	req.URL.RawQuery = q.Encode()
	debugMsg("Request: ", req)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		debugMsg("Error when sending request", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		debugMsg("Error reading response")
		return nil, err
	}

	//fmt.Printf("Code: %v\n", resp.Status)
	debugMsg("Resp: ", string(body))
	end := time.Now()

	testResp := Response{http: *resp, body: body}
	result := &TestResult{Case: testCase, Resp: testResp, Duration: end.Sub(start)}

	exps := expectations(call)
	for _, exp := range exps {
		checkErr := exp.check(testResp)
		if checkErr != nil {
			result.Cause = checkErr
			return result, nil
		}
	}

	err = remember(testResp.bodyAsMap(), call.Remember, rememberMap)
	debugMsg("Remember: ", rememberMap)
	if err != nil {
		debugMsg("Error remember")
		return nil, err
	}

	return result, nil
}

func putRememberedVars(str string, rememberMap map[string]string) string {
	res := str
	for varName, val := range rememberMap {
		placeholder := "{" + varName + "}"
		res = strings.Replace(res, placeholder, val, -1)
	}
	return res
}

func expectations(call Call) (exps []ResponseExpectation) {

	if call.Expect.StatusCode != -1 {
		exps = append(exps, StatusExpectation{statusCode: call.Expect.StatusCode})
	}

	if call.Expect.BodySchema != "" {
		// for now use path relative to suiteDir
		uri := "file:///" + getFileUti(*suiteDir, call.Expect.BodySchema)
		exps = append(exps, BodySchemaExpectation{schemaURI: uri})
	}

	if len(call.Expect.Body) > 0 {
		exps = append(exps, BodyExpectation{pathExpectations: call.Expect.Body})
	}

	if call.Expect.ContentType != "" {
		extractFunc := func(resp http.Response) string {
			contentType, _, _ := mime.ParseMediaType(resp.Header.Get("content-type"))
			return contentType
		}
		exps = append(exps, HeaderExpectation{"content-type", call.Expect.ContentType, extractFunc})
	}

	// and so on
	return exps
}

func getFileUti(dir string, file string) string {
	uri, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println(err)
	}

	uri = filepath.ToSlash(filepath.Join(uri, file))

	return uri
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

// exact value by exact path
func getByPath(m interface{}, path ...string) (string, error) {

	for _, p := range path {
		//fmt.Println(p)
		funcVal, ok := pathFunction(m, p)
		if ok {
			return funcVal, nil
		}

		idx, err := strconv.Atoi(p)
		if err != nil {
			//fmt.Println(err)
			mp, ok := m.(map[string]interface{})
			if !ok {
				str := fmt.Sprintf("Can't cast to Map and get key [%v] in path %v", p, path)
				return "", errors.New(str)
			}
			m = mp[p]
		} else {
			arr, ok := m.([]interface{})
			if !ok {
				str := fmt.Sprintf("Can't cast to Array and get index [%v] in path %v", idx, path)
				return "", errors.New(str)
			}
			if idx >= len(arr) {
				str := fmt.Sprintf("Array only has [%v] elements. Can't get element by index [%v] (counts from zero)", len(arr), idx)
				return "", errors.New(str)
			}
			m = arr[idx]
		}
	}

	if str, ok := castToString(m); ok {
		return str, nil
	} else {
		strErr := fmt.Sprintf("Can't cast path result to string: %v", m)
		return "", errors.New(strErr)
	}
}

// search passing maps and arrays
func searchByPath(m interface{}, s string, path ...string) bool {
	for idx, p := range path {
		//fmt.Println("s ", idx, "p ", p)
		funcVal, ok := pathFunction(m, p)
		if ok {
			if s == funcVal {
				return true
			}
		}

		switch typedM := m.(type) {
		case map[string]interface{}:
			m = typedM[p]
			//fmt.Println("[",m, "] [", s,"]", reflect.TypeOf(m))
			if str, ok := castToString(m); ok {
				if str == s {
					return true
				}
			}
		case []interface{}:
			//fmt.Println("path ", path[idx:])
			for _, obj := range typedM {
				found := searchByPath(obj, s, path[idx:]...)
				if found {
					return true
				}
			}
		}
	}

	return false
}

func castToString(m interface{}) (string, bool) {
	if str, ok := m.(string); ok {
		return str, ok
	} else if flt, ok := m.(float64); ok {
		// numbers (like ids) are parsed as float64 from json
		return strconv.FormatFloat(flt, 'f', 0, 64), ok
	} else {
		return "", ok
	}
}

func pathFunction(m interface{}, pathPart string) (string, bool) {

	if pathPart == "size()" {
		if arr, ok := m.([]interface{}); ok {
			return strconv.Itoa(len(arr)), true
		}
	}

	return "", false
}

type TestResult struct {
	Suite TestSuite
	Case  TestCase
	Resp  Response
	// in case test failed, cause must be specified
	Cause    error
	Duration time.Duration
}

type Response struct {
	http http.Response
	body []byte
}

func (e Response) bodyAsMap() map[string]interface{} {
	var bodyMap map[string]interface{}
	var err error

	contentType, _, _ := mime.ParseMediaType(e.http.Header.Get("content-type"))
	if contentType == "application/xml" {
		err = xml.Unmarshal(e.body, &bodyMap)
	}
	if contentType == "application/json" {
		err = json.Unmarshal(e.body, &bodyMap)
	}

	if err != nil {
		panic(err.Error())
	}

	return bodyMap
}

func debugMsg(a ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Print("\t")
	fmt.Println(a...)
}

type ResponseExpectation interface {
	check(resp Response) error
}

type StatusExpectation struct {
	statusCode int
}

func (e StatusExpectation) check(resp Response) error {
	if resp.http.StatusCode != e.statusCode {
		msg := fmt.Sprintf("Unexpected Status Code. Expected: %d, Actual: %d\n", e.statusCode, resp.http.StatusCode)
		return errors.New(msg)
	}
	return nil
}

type BodySchemaExpectation struct {
	schemaURI string
}

func (e BodySchemaExpectation) check(resp Response) error {
	schemaLoader := gojsonschema.NewReferenceLoader(e.schemaURI)
	documentLoader := gojsonschema.NewStringLoader(string(resp.body))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if !result.Valid() {
		msg := "Unexpected Body Schema:\n"
		for _, desc := range result.Errors() {
			msg = fmt.Sprintf(msg+"%s\n", desc)
		}
		return errors.New(msg)
	}

	return nil
}

type BodyExpectation struct {
	pathExpectations map[string]string
}

func (e BodyExpectation) check(resp Response) error {

	errs := []string{}
	for path, expectedValue := range e.pathExpectations {
		exactMatch := !strings.HasPrefix(path, "~")

		path := strings.Replace(path, "~", "", -1)

		splitPath := strings.Split(path, ".")

		// TODO need rememberedMap here:  expectedValue = putRememberedVars(expectedValue, rememberedMap)
		m := resp.bodyAsMap()

		if (exactMatch) {
			val, err := getByPath(m, splitPath...)
			if val != expectedValue {
				str := fmt.Sprintf("Expected value [%s] on path [%s] does not match [%v].", expectedValue, path, val)
				if err != nil {
					str += " " + err.Error()
				}
				errs = append(errs, str)
			}
		} else {
			found := searchByPath(m, expectedValue, splitPath...)
			if !found {
				err := "Expected value: [" + expectedValue + "] is not found by path: [" + path + "]" // TODO specific message for functions
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		var msg string
		for _, err := range errs {
			msg += err + "\n"
		}
		return errors.New(msg)
	}

	return nil
}

type HeaderExpectation struct {
	headerName  string
	headerValue string
	extractFunc func(http.Response) string
}

func (e HeaderExpectation) check(resp Response) error {
	var value string
	if e.extractFunc == nil {
		value = resp.http.Header.Get(e.headerName)
	} else {
		value = e.extractFunc(resp.http)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("Missing header. Expected \"%s: %s\"\n", e.headerName, e.headerValue)
	}
	if e.headerValue != "" && e.headerValue != value {
		msg := "Unexpected header. Expected \"%s: %s\". Actual \"%s: %s\"\n"
		return fmt.Errorf(msg, e.headerName, e.headerValue, e.headerName, value)
	}
	return nil
}

// TODO add file name to test case report (same names in different files are annoying)
// TODO expect response headers
// TODO separate path and cmd line key for json/xml schema folder
// TODO xml parsing to map (see failing TestXmlUnmarshal)
// TODO add suite.json schema validation to prevent invalid cases (invalid expectation is in file, but never checked)

// optional/under discussion
// TODO "description" in Call for better reporting
// TODO "comment" in test case to describe in more details (sentence)
// TODO matchers: not() ?
// TODO rename remember > keep or memo ?
// TODO full body expectation from file (security testing)
