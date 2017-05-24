package main

import (
	"net/http"
	"testing"
)

func TestResponseBodyOnce(t *testing.T) {
	resp := Response{
		http: http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}},
		},
		body: []byte(`{"key":true}`),
	}

	resp.Body() // first call to parse valid body
	resp.body = []byte(`# set invalid body so it fails if parsed`)

	parsedBody, err := resp.Body()
	if parsedBody == nil || err != nil {
		t.Error("body", parsedBody, "err", err)
	}
}
func TestParseEmptyResponse(t *testing.T) {
	resp := Response{
		body: make([]byte, 0),
		http: http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}},
		},
	}

	data, err := resp.Body()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if data != nil {
		t.Error("Unexpected data.")
	}
}
