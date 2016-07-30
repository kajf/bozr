package main

import (
	"net/http"
	"time"
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
	StatusCode int `json:"statusCode"`
	// shortcut for content-type header
	ContentType string            `json:"contentType"`
	Headers     map[string]string `json:"headers"`
	Body        map[string]string `json:"body"`
	BodySchema  string            `json:"bodySchema"`
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
