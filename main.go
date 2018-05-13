package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	version = "0.8.11"
)

func init() {
	flag.Usage = func() {
		h := "Usage:\n"
		h += "  bozr [OPTIONS] (DIR|FILE)\n\n"

		h += "Options:\n"
		h += "  -d, --debug		Enable debug mode\n"
		h += "  -H, --host		Server to test\n"
		h += "  -w, --worker		Execute in parallel with specified number of workers\n"
		h += "      --throttle	Execute no more than specified number of requests per second (in suite)\n"
		h += "  -h, --help		Print usage\n"
		h += "  -i, --info		Enable info mode. Print request and response details\n"
		h += "      --junit		Enable junit xml reporter\n"
		h += "      --junit-output	Destination for junit report files\n"
		h += "  -v, --version		Print version information and quit\n\n"

		h += "Examples:\n"
		h += "  bozr ./examples\n"
		h += "  bozr -w 2 ./examples\n"
		h += "  bozr -H http://example.com ./examples \n"

		fmt.Fprintf(os.Stderr, h)
	}
}

var (
	suitesDir       string
	hostFlag        string
	workersFlag     int
	throttleFlag    int
	infoFlag        bool
	debugFlag       bool
	helpFlag        bool
	versionFlag     bool
	junitFlag       bool
	junitOutputFlag string

	debug *log.Logger
)

const (
	suiteExt        = ".suite.json"
	ignoredSuiteExt = ".xsuite.json"
)

func initLogger() {
	debugHandler := ioutil.Discard

	if debugFlag {
		debugHandler = os.Stdout
	}

	debug = log.New(debugHandler, "DEBUG: ", log.Ltime|log.Lshortfile)
}

func main() {
	flag.BoolVar(&debugFlag, "d", false, "Enable debug mode.")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug mode")

	flag.BoolVar(&infoFlag, "i", false, "Enable info mode. Print request and response details.")
	flag.BoolVar(&infoFlag, "info", false, "Enable info mode. Print request and response details.")

	flag.StringVar(&hostFlag, "H", "", "Test server address. Example: http://example.com/api.")
	flag.IntVar(&workersFlag, "w", 1, "Execute test sutes in parallel with provided numer of workers. Default is 1.")
	flag.IntVar(&throttleFlag, "throttle", 0, "Execute no more than specified number of requests per second (in suite)")

	flag.BoolVar(&helpFlag, "h", false, "Print usage")
	flag.BoolVar(&helpFlag, "help", false, "Print usage")

	flag.BoolVar(&versionFlag, "v", false, "Print version information and quit")
	flag.BoolVar(&versionFlag, "version", false, "Print version information and quit")

	flag.BoolVar(&junitFlag, "junit", false, "Enable junit xml reporter")
	flag.StringVar(&junitOutputFlag, "junit-output", "./report", "Destination for junit report files. Default ")

	flag.Parse()

	initLogger()

	if versionFlag {
		fmt.Println("bozr version " + version)
		return
	}

	if helpFlag {
		flag.Usage()
		return
	}

	if len(hostFlag) > 0 {
		_, err := url.ParseRequestURI(hostFlag)
		if err != nil {
			terminate("Invalid host is specified.")
			return
		}
	}

	if workersFlag < 1 || workersFlag > 9 {
		fmt.Println("Invalid number of workers:  [", workersFlag, "]. Setting to default [1]")
		workersFlag = 1
	}

	suitesDir = flag.Arg(0)

	if suitesDir == "" {
		flag.Usage()
		fmt.Println()
		terminate("You must specify a directory or file with tests.")
		return
	}

	// check specified source dir/file exists
	_, err := os.Lstat(suitesDir)
	if err != nil {
		terminate(err.Error())
		return
	}

	err = ValidateSuites(suitesDir, suiteExt, ignoredSuiteExt)
	if err != nil {
		terminate("One or more test suites are invalid.", err.Error())
		return
	}

	loader := NewSuiteLoader(suitesDir, suiteExt, ignoredSuiteExt)
	reporter := createReporter()

	RunParallel(loader, reporter, runSuite, workersFlag)
}

func runSuite(suite TestSuite) []TestResult {
	results := []TestResult{}

	throttle := NewThrottle(throttleFlag, time.Second)

	for _, testCase := range suite.Cases {

		result := TestResult{
			Suite:     suite,
			Case:      testCase,
			ExecFrame: TimeFrame{Start: time.Now(), End: time.Now()},
		}

		if testCase.Ignore != nil {
			result.Skipped = true
			result.SkippedMsg = *testCase.Ignore

			results = append(results, result)
			continue
		}

		vars := NewVars()
		for i, c := range testCase.Calls {

			throttle.RunOrPause()

			vars.AddAll(c.Args)

			trace := call(suite.Dir, c, vars)
			trace.Num = i

			result.Traces = append(result.Traces, trace)

			if trace.hasError() {
				break
			}
		}

		result.ExecFrame.End = time.Now()

		results = append(results, result)
	}

	return results
}

func createReporter() Reporter {
	reporters := []Reporter{NewConsoleReporter(infoFlag)}
	if junitFlag {
		path, _ := filepath.Abs(junitOutputFlag)
		reporters = append(reporters, NewJUnitReporter(path))
	}
	reporter := NewMultiReporter(reporters...)
	reporter.Init()

	return reporter
}

func call(suitePath string, call Call, vars *Vars) *CallTrace {

	trace := &CallTrace{}
	execStart := time.Now()

	on := call.On

	bodyContent, err := on.BodyContent(suitePath)
	if err != nil {
		trace.ErrorCause = err
		return trace
	}

	req, err := populateRequest(on, bodyContent, vars)
	if err != nil {
		trace.ErrorCause = err
		return trace
	}

	trace.RequestDump = dumpRequest(req, bodyContent)
	trace.RequestMethod = req.Method
	trace.RequestURL = req.URL.String()

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		debug.Print("Error when sending request", err)
		trace.ErrorCause = err
		return trace
	}

	defer resp.Body.Close()

	trace.ExecFrame = TimeFrame{Start: execStart, End: time.Now()}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		debug.Print("Error reading response")
		trace.ErrorCause = err
		return trace
	}

	testResp := Response{http: resp, body: body}
	trace.ResponseDump = testResp.ToString()

	if err = call.Expect.populateWith(*vars); err != nil {
		trace.ErrorCause = err
		return trace
	}

	exps, err := expectations(call.Expect, suitePath)
	if err != nil {
		trace.ErrorCause = err
		return trace
	}

	for _, exp := range exps {
		checkErr := exp.check(&testResp)

		if checkErr != nil {
			trace.addError(checkErr)
			return trace
		}

		trace.addExp(exp.desc())
	}

	err = rememberBody(&testResp, call.Remember.Body, vars)
	debug.Print("Remember: ", vars)
	if err != nil {
		debug.Print("Error remember")
		trace.ErrorCause = err
		return trace
	}

	rememberHeaders(testResp.http.Header, call.Remember.Headers, vars)

	return trace
}

func populateRequest(on On, bodyTmpl string, vars *Vars) (*http.Request, error) {

	urlStr, err := urlPrefix(vars.ApplyTo(on.URL))
	if err != nil {
		return nil, errors.New("Cannot create request. Invalid url: " + on.URL)
	}

	ctx := NewTemplateContext(vars)

	body, err := executeTemplate(ctx, bodyTmpl)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse test body: %s", err.Error())
	}

	dat := []byte(body)

	req, err := http.NewRequest(on.Method, urlStr, bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	for key, valueTmpl := range on.Headers {
		value, err := executeTemplate(ctx, valueTmpl)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse header value: %s", err.Error())
		}

		req.Header.Add(key, value)
	}

	q := req.URL.Query()
	for key, valueTmpl := range on.Params {
		value, err := executeTemplate(ctx, valueTmpl)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse query param value: %s", err.Error())
		}

		q.Add(key, value)
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
	baseURL, err := url.ParseRequestURI(base)
	if err != nil {
		return "", err
	}
	return baseURL.Scheme + "://" + baseURL.Host + path.Join(baseURL.Path, p), nil
}

func expectations(expect Expect, suitePath string) ([]ResponseExpectation, error) {
	var exps []ResponseExpectation
	if expect.StatusCode != 0 {
		exps = append(exps, StatusCodeExpectation{statusCode: expect.StatusCode})
	}

	if expect.HasSchema() {

		schemeURI, err := expect.BodySchema(suitePath)
		if err != nil {
			return nil, err
		}

		exps = append(exps, BodySchemaExpectation{schemaURI: schemeURI})
	}

	if len(expect.Body) > 0 {
		exps = append(exps, BodyExpectation{pathExpectations: expect.Body})
	}

	if len(expect.Absent) > 0 {
		exps = append(exps, AbsentExpectation{paths: expect.Absent})
	}

	if len(expect.Headers) > 0 {
		for k, v := range expect.Headers {
			exps = append(exps, HeaderExpectation{Name: k, Value: v})
		}
	}

	if expect.ContentType != "" {
		exps = append(exps, ContentTypeExpectation{expect.ContentType})
	}

	// and so on
	return exps, nil
}

func rememberBody(resp *Response, remember map[string]string, vars *Vars) (err error) {

	for varName, pathLine := range remember {
		body, err := resp.Body()
		if err != nil {
			debug.Print("Can't parse response body to Map for [remember]")
			return err
		}

		if rememberVar, err := GetByPath(body, pathLine); err == nil {
			vars.Add(varName, rememberVar)
		} else {
			debug.Print(err)
			return fmt.Errorf("Remembered value not found, path: %v", pathLine)
		}
	}

	return err
}

func rememberHeaders(header http.Header, remember map[string]string, vars *Vars) {
	for valueName, headerName := range remember {
		value := header.Get(headerName)
		if value == "" {
			continue
		}

		vars.Add(valueName, value)
	}
}

func dumpRequest(req *http.Request, body string) string {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s %s %s\n", req.Method, req.URL.String(), req.Proto))

	for k, v := range req.Header {
		buf.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, " ")))
	}

	if len(body) > 0 {
		buf.WriteString("\n")
		buf.WriteString(body)
	}

	return buf.String()
}

func terminate(msgLines ...string) {
	for _, line := range msgLines {
		fmt.Fprintln(os.Stderr, line)
	}

	os.Exit(1)
}
