package main

import (
	"net/http"
	"testing"
)

func TestExpectedStatusCode(t *testing.T) {
	exp := StatusCodeExpectation{statusCode: 200}
	err := exp.check(Response{
		http: http.Response{
			StatusCode: 400,
		},
		body: nil,
	})

	if err == nil {
		t.Fail()
	}
}

func TestUnexpectedStatusCode(t *testing.T) {
	exp := StatusCodeExpectation{statusCode: 200}
	err := exp.check(Response{
		http: http.Response{
			StatusCode: 200,
		},
		body: nil,
	})

	if err != nil {
		t.Fail()
	}

}

func TestExpectedHeader(t *testing.T) {
	exp := HeaderExpectation{Name: "X-Test", Value: "PASS"}

	err := exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"X-Test": []string{"PASS"}},
		},
		body: nil,
	})

	if err != nil {
		t.Fail()
	}
}

func TestUnexpectedHeader(t *testing.T) {
	exp := HeaderExpectation{Name: "X-Test", Value: "PASS"}

	err := exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"X-Test": []string{"FAILED"}},
		},
		body: nil,
	})

	if err == nil {
		t.Fail()
	}
}

func TestExpectedContentType(t *testing.T) {
	exp := ContentTypeExpectation{Value: "application/json"}

	err := exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json;charset=utf-8"}},
		},
		body: nil,
	})

	if err != nil {
		t.Error(err.Error())
	}
}

func TestUnexpectedContentType(t *testing.T) {
	exp := ContentTypeExpectation{Value: "application/json"}

	err := exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"text/html; charset=utf-8"}},
		},
		body: nil,
	})

	if err == nil {
		t.Fail()
	}
}
