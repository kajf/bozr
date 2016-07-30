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
	"strconv"
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

	debugMsgF("Process file: %s\n", info.Name())
	content, e := ioutil.ReadFile(path)

	if e != nil {
		debugMsgF("File error: %v\n", e)
		return nil
	}

	var testCases []TestCase
	err = json.Unmarshal(content, &testCases)
	if err != nil {
		debugMsgF("Parse error: %v\n", err)
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
		uri := getFileUri(*suiteDir, on.BodyFile)
		if d, err := ioutil.ReadFile(uri); err == nil {
			dat = d
		} else {
			debugMsg("Can't read body file: ", err.Error())
		}
	}

	req, _ := http.NewRequest(on.Method, *host+on.URL, bytes.NewBuffer(dat))

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

	m, err := testResp.bodyAsMap()
	if err != nil {
		debugMsg("Can't parse response body to Map for [Remember]")
		return nil, err
	}

	err = remember(m, call.Remember, rememberMap)
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
		uri := getFileUri(*suiteDir, call.Expect.BodySchema)
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

func getFileUri(dir string, file string) string {
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

func (e Response) bodyAsMap() (map[string]interface{}, error) {
	var bodyMap map[string]interface{}
	var err error

	contentType, _, _ := mime.ParseMediaType(e.http.Header.Get("content-type"))
	if contentType == "application/xml" {
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

// TODO separate path and cmd line key for json/xml schema folder
// TODO add suite.json schema validation to prevent invalid cases (invalid expectation is in file, but never checked)

// optional/under discussion
// TODO "description" in Call for better reporting
// TODO matchers: not() ?
// TODO rename remember > keep or memo ?
// TODO full body expectation from file (security testing)
