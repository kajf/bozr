package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestRememberBodyLazy(t *testing.T) {
	resp := Response{
		http: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}},
		},
		body: []byte(`# invalid body so it fails if parsed`),
	}

	err := rememberBody(&resp, map[string]string{}, NewVars(""))

	if err != nil {
		t.Error(err)
	}
}

func TestRunSuite_InvalidUrlAndUnused_OnlyInvalidUrl(t *testing.T) {
	suite := TestSuite{
		Cases: []TestCase{
			{
				Calls: []Call{
					{
						Args: map[string]interface{}{"a": 1},
						On: On{
							URL: "my-invalid-host",
						},
					},
				},
			},
		},
	}

	results := runSuite(suite)

	err := results[0].Traces[0].ErrorCause
	if err == nil || !strings.Contains(err.Error(), "Invalid url") {
		t.Error("Expected error not thrown", err)
	}
}

func TestConcatURL(t *testing.T) {

	t.Run("open base and closed path", func(t *testing.T) {
		base := "http://example.com"
		path := "/api/v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

	t.Run("closed base and closed path", func(t *testing.T) {
		base := "http://example.com/api/"
		path := "/v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

	t.Run("closed base and open path", func(t *testing.T) {
		base := "http://example.com/api/"
		path := "v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

	t.Run("open base and open path", func(t *testing.T) {
		base := "http://example.com/api"
		path := "v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

}
