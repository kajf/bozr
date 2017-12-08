package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
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

func (suite TestSuite) PackageName() string {
	if strings.HasPrefix(suite.Dir, ".") {
		return ""
	}

	return strings.Replace(filepath.ToSlash(suite.Dir), "/", ".", -1)
}

func (suite TestSuite) FullName() string {
	pkg := suite.PackageName()
	if pkg == "" {
		return suite.Name
	}

	return fmt.Sprintf("%s.%s", suite.PackageName(), suite.Name)
}

type TestCase struct {
	Name   string  `json:"name,omitempty"`
	Ignore *string `json:"ignore,omitempty"`
	Calls  []Call  `json:"calls,omitempty"`
}

type Call struct {
	Args     map[string]interface{} `json:"args,omitempty"`
	On       On                     `json:"on,omitempty"`
	Expect   Expect                 `json:"expect,omitempty"`
	Remember Remember               `json:"remember,omitempty"`
}

type Remember struct {
	Body    map[string]string `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type On struct {
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Params   map[string]string `json:"params"`
	Body     json.RawMessage   `json:"body"`
	BodyFile string            `json:"bodyFile"`
}

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

type TimeFrame struct {
	Start time.Time
	End   time.Time
}

func (tf TimeFrame) Duration() time.Duration {
	return tf.End.Sub(tf.Start)
}

func (tf *TimeFrame) Extend(tf2 TimeFrame) {
	if tf.Start.After(tf2.Start) {
		tf.Start = tf2.Start
	}

	if tf.End.Before(tf2.End) {
		tf.End = tf2.End
	}
}

type TError struct {
	Resp  Response
	Cause error
}

// Response wraps test call http response
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
		body = resp.body
	}

	details := fmt.Sprintf("%s \n %s \n%s", http.Status, headers, body)
	return details
}

// Throttle implements rate limiting based on sliding time window
type Throttle struct {
	limit     int
	timeFrame time.Duration
	queue     []time.Time
}

// NewThrottle creates Throttle with following notation: not more than X executions per time period
// e.g. not more than 300 calls per 1 minute
// zero limit means no limit
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
	if t.limit == 0 {
		return
	} // no limit, so exit

	t.cleanOld()

	totlalCallsInFrame := len(t.queue)
	limitExceeded := (totlalCallsInFrame == t.limit)

	if limitExceeded {
		eldestCallInFrame := t.queue[0]
		durationSinceEldest := time.Since(eldestCallInFrame)

		remaining := t.timeFrame - durationSinceEldest

		time.Sleep(remaining)

		t.queue = t.queue[1:] // free up space for new item
	}

	t.queue = append(t.queue, time.Now())
}
