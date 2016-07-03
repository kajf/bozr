package main

import (
	"testing"
	"encoding/json"
)

func TestGetByPathSimple(t *testing.T) {
	token := "abc"
	m := map[string]interface{}{"token": token, "bar": 2}

	got := getByPath(m, "token")

	if got != token {
		t.Error(
			"For", "token",
			"expected", token,
			"got", got,
		)
	}
}

func TestGetByPath2ndLevel(t *testing.T) {
	name := "abc"
	token := map[string]interface{}{"name": name}
	m := map[string]interface{}{"token": token, "bar": 2}

	got := getByPath(m, "token", "name")

	if got != name {
		t.Error(
			"For", "token.name",
			"expected", name,
			"got", got,
		)
	}
}

func TestGetByPathWithIndex(t *testing.T) {
	s := `{
		"items":[
			{"id":"417857","status":"OK"},
			{"id":"417858","status":"OK"}
		]
	 	}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	got := getByPath(m, "items", "0", "id")
	if got != "417857" {
		t.Error(
			"expected", "417857",
			"got", got,
		)
	}
}

func TestGetByPathEmpty(t *testing.T) {
	emptyMap := make(map[string]interface{})

	got := getByPath(emptyMap, "token")

	if got != nil {
		t.Error(
			"For", "token",
			"expected", nil,
			"got", got,
		)
	}
}

func TestPutRememberedVars(t *testing.T) {
	token := "test_token"
	rememberMap := map[string]string{"savedToken":token}

	got := putRememberedVars("bearer {savedToken}", rememberMap)

	if got != "bearer " + token {
		t.Error(
			"expected", "bearer " + token,
			"got", got,
		)
	}
}

func TestPutRememberedVarsMultiple(t *testing.T) {
	token := "test_token"
	second := "second"
	rememberMap := map[string]string{"savedToken":token, "aSecond":second}

	got := putRememberedVars("prefix {savedToken} middle {aSecond} postfix", rememberMap)

	expected := "prefix " + token + " middle " + second +" postfix"
	if got != expected {
		t.Error(
			"expected[", expected,
			"got[", got,
		)
	}
}
