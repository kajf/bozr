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
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
		},
		body: nil,
	})

	if err != nil {
		t.Error(err.Error())
	}
}

func TestExpectedContentTypeIgnoreEncoding(t *testing.T) {
	exp := ContentTypeExpectation{Value: "application/json"}

	err := exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json; charset=utf-8"}},
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
			Header: map[string][]string{"Content-Type": []string{"text/html"}},
		},
		body: nil,
	})

	if err == nil {
		t.Fail()
	}
}

func TestBodyExpectationBool(t *testing.T) {
	m, err := jsonAsMap(`{
		"flag": true
	 	}`)
	if err != nil {
		t.Error(err)
	}

	exp := BodyExpectation{pathExpectations: m}

	err = exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
		},
		body: []byte("{\"flag\":true}"),
	})

	if err != nil {
		t.Error(err)
	}
}

func TestBodyExpectationInt(t *testing.T) {
	m, err := jsonAsMap(`{
		"len": 2
	 	}`)
	if err != nil {
		t.Error(err)
	}

	exp := BodyExpectation{pathExpectations: m}

	err = exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
		},
		body: []byte("{\"len\":2}"),
	})

	if err != nil {
		t.Error(err)
	}
}

func TestBodyExpectationSize(t *testing.T) {
	m, err := jsonAsMap(`{
		"items.size()": 0
	 	}`)
	if err != nil {
		t.Error(err)
	}

	exp := BodyExpectation{pathExpectations: m}

	err = exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
		},
		body: []byte("{\"items\":[]}"),
	})

	if err != nil {
		t.Error(err)
	}
}

func TestBodyExpectationSearchBool(t *testing.T) {
	m, err := jsonAsMap(`{
		"~flag": true
	 	}`)
	if err != nil {
		t.Error(err)
	}

	exp := BodyExpectation{pathExpectations: m}

	err = exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
		},
		body: []byte("{\"flag\":true}"),
	})

	if err != nil {
		t.Error(err)
	}
}

func TestBodyExpectationSearchInt(t *testing.T) {
	m, err := jsonAsMap(`{
		"~len": 2
	 	}`)
	if err != nil {
		t.Error(err)
	}

	exp := BodyExpectation{pathExpectations: m}

	err = exp.check(Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
		},
		body: []byte("{\"len\":2}"),
	})

	if err != nil {
		t.Error(err)
	}
}
