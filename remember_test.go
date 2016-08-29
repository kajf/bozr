package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPopulateRequestBody(t *testing.T) {
	//given
	on := On{URL: "http://example.com"}
	value := "abc"

	// when
	req, _ := populateRequest(on, "pre {var} post", map[string]interface{}{"var": value})

	//then
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	got := buf.String()
	if !strings.Contains(got, value) {
		t.Error(
			"body does not conatain value", value,
			"got", got,
		)
	}
}

func TestPopulateRememberedVars(t *testing.T) {
	token := "test_token"
	rememberMap := map[string]interface{}{"savedToken": token}

	got := populateRememberedVars("bearer {savedToken}", rememberMap)

	if got != "bearer "+token {
		t.Error(
			"expected", "bearer "+token,
			"got", got,
		)
	}
}

func TestPopulateRememberedVarsMultiple(t *testing.T) {
	token := "test_token"
	second := "second"
	rememberMap := map[string]interface{}{"savedToken": token, "aSecond": second}

	got := populateRememberedVars("prefix {savedToken} middle {aSecond} postfix", rememberMap)

	expected := "prefix " + token + " middle " + second + " postfix"
	if got != expected {
		t.Error(
			"expected[", expected,
			"got[", got,
		)
	}
}
