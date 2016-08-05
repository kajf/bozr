package main

import (
	"testing"
	"bytes"
	"strings"
)

func TestPopulateRequestBody(t *testing.T) {
	//given
	on := On{}
	value := "abc"

	// when
	req := populateRequest(on, "pre {var} post", map[string]string{"var": value})

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
	rememberMap := map[string]string{"savedToken":token}

	got := populateRememberedVars("bearer {savedToken}", rememberMap)

	if got != "bearer " + token {
		t.Error(
			"expected", "bearer " + token,
			"got", got,
		)
	}
}

func TestPopulateRememberedVarsMultiple(t *testing.T) {
	token := "test_token"
	second := "second"
	rememberMap := map[string]string{"savedToken":token, "aSecond":second}

	got := populateRememberedVars("prefix {savedToken} middle {aSecond} postfix", rememberMap)

	expected := "prefix " + token + " middle " + second +" postfix"
	if got != expected {
		t.Error(
			"expected[", expected,
			"got[", got,
		)
	}
}
