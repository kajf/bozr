package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
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

type TestCase struct {
	Name  string `json:"name"`
	Calls []Call `json:"calls"`
}

type Call struct {
	Args     map[string]interface{} `json:"args"`
	On       On                     `json:"on"`
	Expect   Expect                 `json:"expect"`
	Remember map[string]string      `json:"remember"`
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
	BodySchemaFile string                 `json:"bodySchemaFile"`
	BodySchemaURI  string                 `json:"bodySchemaURI"`
}

func (e Expect) hasSchema() bool {
	return e.BodySchemaFile != "" || e.BodySchemaURI != ""
}

type TestResult struct {
	Suite TestSuite
	Case  TestCase
	// in case test failed, cause must be specified
	Error    *TError
	Duration time.Duration
}

type TError struct {
	Resp  Response
	Cause error
}

type Response struct {
	http http.Response
	body []byte
}

func (e Response) parseBody() (interface{}, error) {
	contentType, _, _ := mime.ParseMediaType(e.http.Header.Get("content-type"))
	if contentType == "application/xml" || contentType == "text/xml" {
		m, err := mxj.NewMapXml(e.body)
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
		if string(e.body[0]) == "[" {
			body = make([]interface{}, 0)
			err = json.Unmarshal(e.body, &body)
		} else {
			body = make(map[string]interface{})
			err = json.Unmarshal(e.body, &body)
		}

		if err == nil {
			return body, nil
		}
		return err, nil
	}

	return nil, errors.New("Cannot parse body. Unsupported content type")
}

// ToString return string representation of response data
// including status code, headers and body.
func (resp Response) ToString() string {
	http := resp.http

	headers := "\n"
	for k, v := range http.Header {
		headers = fmt.Sprintf("%s%s: %s\n", headers, k, strings.Join(v, " "))
	}

	var body interface{}
	contentType, _, _ := mime.ParseMediaType(resp.http.Header.Get("content-type"))
	if contentType == "application/json" {
		data, _ := resp.parseBody()
		body, _ = json.MarshalIndent(data, "", "  ")
	}

	if contentType == "application/xml" || contentType == "text/xml" {
		resp.parseBody()
		mp, _ := mxj.NewMapXml(resp.body, false)
		body, _ = mp.XmlIndent("", "  ")
	}

	if body == nil {
		body = resp.body
	}

	details := fmt.Sprintf("%s \n %s \n%s", http.Status, headers, body)
	return details
}
