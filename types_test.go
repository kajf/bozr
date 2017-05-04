package main

import (
	"net/http"
	"testing"
)

func TestParseEmptyResponse(t *testing.T) {
	resp := Response{
		body: make([]byte, 0),
		http: http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}},
		},
	}

	data, err := resp.parseBody()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if data != nil {
		t.Error("Unexpected data.")
	}
}

func TestBodyShortedRemember(t *testing.T) {
	call := Call{
		RawRemember: map[string]interface{}{
			"key": "path.to.remember",
		},
	}

	remember := call.Remember(RememberSourceBody)

	if remember["key"] != "path.to.remember" {
		t.Errorf("Unexpected path in the body to remember. Actual: %v", remember)
	}
}

func TestBodyRemember(t *testing.T) {
	call := Call{
		RawRemember: map[string]interface{}{
			"secret_key": map[string]interface{}{
				"body": "path.to.remember",
			},
		},
	}

	remember := call.Remember(RememberSourceBody)

	if remember["secret_key"] != "path.to.remember" {
		t.Errorf("Unexpected path in the body to remember. Actual: %v", remember)
	}
}

func TestHeaderRemember(t *testing.T) {
	call := Call{
		RawRemember: map[string]interface{}{
			"loc": map[string]interface{}{
				"header": "Location",
			},
		},
	}

	remember := call.Remember(RememberSourceHeader)

	if remember["loc"] != "Location" {
		t.Errorf("Unexpected header to remember. Actual: %v", remember)
	}
}
