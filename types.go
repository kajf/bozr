package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clbanning/mxj"
)

// TestSuite represents file with test cases.
type TestSuite struct {
	// file name
	Name string
	// Path to a directory where suite is located
	// Relative to the suite root
	Dir string
	// test cases listed in a file
	Cases []TestCase
}

// PackageName builds name of a package based on folder where test is located
func (suite TestSuite) PackageName() string {
	if strings.HasPrefix(suite.Dir, ".") {
		return ""
	}

	return strings.Replace(filepath.ToSlash(suite.Dir), "/", ".", -1)
}

// FullName builds name of the test including package and test name
func (suite TestSuite) FullName() string {
	pkg := suite.PackageName()
	if pkg == "" {
		return suite.Name
	}

	return fmt.Sprintf("%s.%s", suite.PackageName(), suite.Name)
}

// TestCase represents single test scenario
type TestCase struct {
	Name   string  `json:"name,omitempty"`
	Ignore *string `json:"ignore,omitempty"`
	Calls  []Call  `json:"calls,omitempty"`
}

// Call defines metadata for one request-response virifiation within TestCase
type Call struct {
	Args     map[string]interface{} `json:"args,omitempty"`
	On       On                     `json:"on,omitempty"`
	Expect   Expect                 `json:"expect,omitempty"`
	Remember Remember               `json:"remember,omitempty"`
}

// Remember defines items from HTTP response to persist for usage in future calls
type Remember struct {
	Body    map[string]string `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// On is a metadata for building a HTTP request
type On struct {
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Params   map[string]string `json:"params"`
	Body     json.RawMessage   `json:"body"`
	BodyFile string            `json:"bodyFile"`
}

// Expect is a metadata for HTTP response verification
type Expect struct {
	StatusCode int `json:"statusCode"`
	// shortcut for content-type header
	ContentType    string                 `json:"contentType"`
	Headers        map[string]string      `json:"headers"`
	Body           map[string]interface{} `json:"body"`
	Absent         []string               `json:"absent"`
	BodySchemaFile string                 `json:"bodySchemaFile"`
	BodySchemaURI  string                 `json:"bodySchemaURI"`
}

func (e Expect) hasSchema() bool {
	return e.BodySchemaFile != "" || e.BodySchemaURI != ""
}

func (e Expect) populateWith(vars Vars) {
	//expect.Headers        map[string]string
	for name, val := range e.Headers {
		e.Headers[name] = vars.ApplyTo(val)
	}

	//expect.Body           map[string]interface{} - string, array, num
	for path, val := range e.Body {

		switch typedExpect := val.(type) {
		case []string:
			for i, el := range typedExpect {
				typedExpect[i] = vars.ApplyTo(el)
			}
		case string:
			e.Body[path] = vars.ApplyTo(typedExpect)
		default:
			// do nothing with values like numbers
		}
	}
}

// TestResult represents single test case for reporting
type TestResult struct {
	Suite      TestSuite
	Case       TestCase
	Skipped    bool
	SkippedMsg string
	// in case test failed, cause must be specified
	Error *TError

	ExecFrame TimeFrame
}

// TimeFrame describes period of time
type TimeFrame struct {
	Start time.Time
	End   time.Time
}

// Duration of TimeFrame from Start to End
func (tf TimeFrame) Duration() time.Duration {
	return tf.End.Sub(tf.Start)
}

// Extend does extension of time perod by other provided TimeFrame
// from earlier start to elder end
func (tf *TimeFrame) Extend(tf2 TimeFrame) {
	if tf.Start.After(tf2.Start) {
		tf.Start = tf2.Start
	}

	if tf.End.Before(tf2.End) {
		tf.End = tf2.End
	}
}

// TError stands for test error in report
type TError struct {
	CallNum int
	Resp    Response
	Cause   error
}

// Response wraps test call HTTP response
type Response struct {
	http       http.Response
	body       []byte
	parsedBody interface{}
}

// Body retruns parsed response (array or map) depending on provided 'Content-Type'
// supported content types are 'application/json', 'application/xml', 'text/xml'
func (resp *Response) Body() (interface{}, error) {
	if resp.parsedBody != nil {
		return resp.parsedBody, nil
	}

	var err error
	resp.parsedBody, err = resp.parseBody()

	return resp.parsedBody, err
}

func (resp Response) parseBody() (interface{}, error) {

	if len(resp.body) == 0 {
		return nil, nil
	}

	contentType, _, _ := mime.ParseMediaType(resp.http.Header.Get("content-type"))
	if contentType == "application/xml" || contentType == "text/xml" {
		m, err := mxj.NewMapXml(resp.body)
		if err == nil {
			return m.Old(), nil
		}
		return nil, err
	}

	if contentType == "application/json" {
		var (
			body interface{}
			err  error
		)
		if string(resp.body[0]) == "[" {
			body = make([]interface{}, 0)
			err = json.Unmarshal(resp.body, &body)
		} else {
			body = make(map[string]interface{})
			err = json.Unmarshal(resp.body, &body)
		}

		if err == nil {
			return body, nil
		}
		return nil, err
	}

	return nil, errors.New("Cannot parse body. Unsupported content type")
}

// ToString return string representation of response data
// including status code, headers and body.
func (resp *Response) ToString() string {
	http := resp.http

	headers := "\n"
	for k, v := range http.Header {
		headers = fmt.Sprintf("%s%s: %s\n", headers, k, strings.Join(v, " "))
	}

	var body interface{}
	contentType, _, _ := mime.ParseMediaType(resp.http.Header.Get("content-type"))
	if contentType == "application/json" {
		data, _ := resp.Body()
		body, _ = json.MarshalIndent(data, "", "  ")
	}

	if contentType == "application/xml" || contentType == "text/xml" {
		resp.Body()
		mp, _ := mxj.NewMapXml(resp.body, false)
		body, _ = mp.XmlIndent("", "  ")
	}

	if body == nil {
		body = ""
	}

	details := fmt.Sprintf("%s \n %s \n%s", http.Status, headers, body)
	return details
}

// Vars defines map of test case level variables (e.g. args, remember, env)
type Vars struct {
	items map[string]interface{}
}

// NewVars create new Vars object with default set of env variables
func NewVars() *Vars {
	v := &Vars{items: make(map[string]interface{})}

	v.addEnv()

	return v
}

func (v *Vars) addEnv() {

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		v.items["env:"+pair[0]] = pair[1]
	}
}

// Add is adding variable with name and value to map
func (v *Vars) Add(name string, val interface{}) {
	v.items[name] = val
}

// AddAll is a shortcut for adding provided map of variables in for-loop
func (v *Vars) AddAll(src map[string]interface{}) {
	for key, val := range src {
		v.items[key] = val
	}
}

// ApplyTo updates input template with values correspondent to placeholders
// according to current vars map
func (v *Vars) ApplyTo(str string) string {
	res := str
	for varName, val := range v.items {
		placeholder := "{" + varName + "}"
		res = strings.Replace(res, placeholder, toString(val), -1)
	}
	return res
}

// Throttle implements rate limiting based on sliding time window
type Throttle struct {
	limit     int
	timeFrame time.Duration
	queue     []time.Time
}

// InfiniteLimit is a constant that represents an absence of any limits.
const InfiniteLimit = 0

// NewThrottle creates Throttle with following notation: not more than X executions per time period
// e.g. not more than 300 calls per 1 minute
func NewThrottle(limit int, perTimeFrame time.Duration) *Throttle {
	return &Throttle{limit: limit, timeFrame: perTimeFrame, queue: make([]time.Time, 0)}
}

func (t *Throttle) cleanOld() {
	for _, callTime := range t.queue {

		timeSince := time.Since(callTime)
		if timeSince <= t.timeFrame {
			break
		} // queue is ordered, so no point to proceed

		t.queue = t.queue[1:]
	} // clean up top callTimes older than frame
}

// RunOrPause should be added to any throttled operation
// so it either runs without interruption or waits for next time frame if current time frame call limit is exceeded
func (t *Throttle) RunOrPause() {
	if t.limit == InfiniteLimit {
		return
	} // no limit, so exit

	t.cleanOld()

	totalCallsInFrame := len(t.queue)
	limitExceeded := (totalCallsInFrame == t.limit)

	if limitExceeded {
		eldestCallInFrame := t.queue[0]
		durationSinceEldest := time.Since(eldestCallInFrame)

		remaining := t.timeFrame - durationSinceEldest

		time.Sleep(remaining)

		t.queue = t.queue[1:] // free up space for new item
	}

	t.queue = append(t.queue, time.Now())
}
