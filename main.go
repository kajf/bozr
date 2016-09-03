package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/clbanning/mxj"
)

const (
	version = "0.8.2"
)

func init() {
	flag.Usage = func() {
		h := "Usage:\n"
		h += "  bozr [OPTIONS] (DIR|FILE)\n\n"

		h += "Options:\n"
		h += "  -d, --debug		Enable debug mode\n"
		h += "  -H, --host		Server to test\n"
		h += "  -h, --help		Print usage\n"
		h += "      --junit		Enable junit xml reporter\n"
		h += "  -v, --version		Print version information and quit\n\n"

		h += "Examples:\n"
		h += "  bozr ./examples\n"
		h += "  bozr -H http://example.com ./examples \n"

		fmt.Fprintf(os.Stderr, h)
	}
}

var (
	suiteDir    string
	hostFlag    string
	debugFlag   bool
	helpFlag    bool
	versionFlag bool
	junitFlag   bool
)

func main() {
	flag.BoolVar(&debugFlag, "d", false, "Enable debug mode")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug mode")

	flag.StringVar(&hostFlag, "H", "http://localhost:8080", "Test server address")

	flag.BoolVar(&helpFlag, "h", false, "Print usage")
	flag.BoolVar(&helpFlag, "help", false, "Print usage")

	flag.BoolVar(&versionFlag, "v", false, "Print version information and quit")
	flag.BoolVar(&versionFlag, "version", false, "Print version information and quit")

	flag.BoolVar(&junitFlag, "junit", false, "Enable junit xml reporter")

	flag.Parse()

	if versionFlag {
		fmt.Println("bozr version " + version)
		return
	}

	if helpFlag {
		flag.Usage()
		return
	}

	src := flag.Arg(0)

	if src == "" {
		fmt.Print("You must specify a directory or file with tests.\n\n")
		flag.Usage()
		return
	}

	var ch <-chan TestSuite
	if filepath.Ext(src) == "" {
		debugMsg("Loading from directory...")
		suiteDir = src
		ch = NewDirLoader(suiteDir)
	} else {
		debugMsg("Loading from file...")
		suiteDir = filepath.Dir(src)
		ch = NewFileLoader(src)
	}

	reporters := []Reporter{NewConsoleReporter()}
	if junitFlag {
		path, _ := filepath.Abs("./report")
		reporters = append(reporters, NewJUnitReporter(path))
	}
	reporter := NewMultiReporter(reporters...)

	// test case runner?
	for suite := range ch {
		for _, testCase := range suite.Cases {

			rememberedMap := make(map[string]interface{})

			for _, c := range testCase.Calls {
				addAll(c.Args, rememberedMap)
				tr := call(suite, testCase, c, rememberedMap)
				tr.Suite = suite
				reporter.Report(*tr)
			}
		}
	}

	reporter.Flush()
}

func addAll(src, target map[string]interface{}) {
	for key, val := range src {
		target[key] = val
	}
}

func call(testSuite TestSuite, testCase TestCase, call Call, rememberMap map[string]interface{}) (result *TestResult) {
	debugMsg("--- Starting call ...") // TODO add call description
	start := time.Now()
	result = &TestResult{Case: testCase}

	on := call.On

	dat := []byte(on.Body)
	if on.BodyFile != "" {
		uri, err := toAbsPath(testSuite.Dir, on.BodyFile)
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

	req, err := populateRequest(on, string(dat), rememberMap)
	if err != nil {
		result.Cause = err
		return
	}
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

	end := time.Now()

	testResp := Response{http: *resp, body: body}
	result.Resp = testResp
	result.Duration = end.Sub(start)

	debugMsg(testResp.ToString())

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

func populateRequest(on On, body string, rememberMap map[string]interface{}) (*http.Request, error) {

	url, err := urlPrefix(populateRememberedVars(on.URL, rememberMap))
	if err != nil {
		return nil, errors.New("Cannot create request. Invalid url: " + on.URL)
	}

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

	return req, nil
}

func urlPrefix(p string) (string, error) {
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p, nil
	}

	return concatURL(hostFlag, p)
}

func concatURL(base string, p string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	return baseURL.Scheme + "://" + baseURL.Host + path.Join(baseURL.Path, p), nil
}

func populateRememberedVars(str string, rememberMap map[string]interface{}) string {
	res := str
	for varName, val := range rememberMap {
		placeholder := "{" + varName + "}"
		res = strings.Replace(res, placeholder, toString(val), -1)
	}
	return res
}

// toString returns value suitable to insert as an argument
// if value if a float where decimal part is zero - convert to int
func toString(rw interface{}) string {
	var sv interface{} = rw
	if fv, ok := rw.(float64); ok {
		_, frac := math.Modf(fv)
		if frac == 0 {
			sv = int(fv)
		}
	}

	return fmt.Sprintf("%v", sv)
}

func expectations(call Call, srcDir string) ([]ResponseExpectation, error) {
	var exps []ResponseExpectation
	if call.Expect.StatusCode != 0 {
		exps = append(exps, StatusCodeExpectation{statusCode: call.Expect.StatusCode})
	}

	if call.Expect.hasSchema() {
		var (
			schemeURI string
			err       error
		)

		if call.Expect.BodySchemaFile != "" {
			schemeURI, err = toAbsPath(srcDir, call.Expect.BodySchemaFile)
			if err != nil {
				return nil, err
			}
			schemeURI = "file:///" + schemeURI
		}

		if call.Expect.BodySchemaURI != "" {
			isHTTP := strings.HasPrefix(call.Expect.BodySchemaURI, "http://")
			isHTTPS := strings.HasPrefix(call.Expect.BodySchemaURI, "https://")
			if !(isHTTP || isHTTPS) {
				schemeURI = hostFlag + call.Expect.BodySchemaURI
			} else {
				schemeURI = call.Expect.BodySchemaURI
			}
		}
		exps = append(exps, BodySchemaExpectation{schemaURI: schemeURI})
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
		exps = append(exps, ContentTypeExpectation{call.Expect.ContentType})
	}

	// and so on
	return exps, nil
}

func toAbsPath(srcDir string, assetPath string) (string, error) {
	if filepath.IsAbs(assetPath) {
		// ignore srcDir
		return assetPath, nil
	}

	uri, err := filepath.Abs(filepath.Join(suiteDir, srcDir, assetPath))
	if err != nil {
		return "", errors.New("Invalid file path: " + assetPath)
	}

	return filepath.ToSlash(uri), nil
}

func remember(bodyMap map[string]interface{}, remember map[string]string, rememberedMap map[string]interface{}) (err error) {

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
	if !debugFlag {
		return
	}
	fmt.Println(a...)
}

func debugMsgF(tpl string, a ...interface{}) {
	if !debugFlag {
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
