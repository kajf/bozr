package main

import (
	"testing"
	"encoding/json"
)

func TestGetByPathSimple(t *testing.T) {
	token := "abc"

	s := `{"token":"` + token + `","ttl":3600000,"units":"milliseconds"}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	got, _ := getByPath(m, "token")

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

	got, _ := getByPath(m, "token", "name")

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

	got, _ := getByPath(m, "items", "0", "id")
	if got != "417857" {
		t.Error(
			"expected", "417857",
			"got", got,
		)
	}
}

func TestGetByPathArraySize(t *testing.T) {
	s := `{
		"items":[
			{"status":"OK"},
			{"status":"OK"}
		]
	 	}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items", "size()")
	if got != "2" || err != nil {
		t.Error(
			"expected 2",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathArrayOutOfBounds(t *testing.T) {
	s := `{
		"items":[
			{"id":"-1","status":"OK"},
			{"id":"-2","status":"OK"}
		]
	 	}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items", "2", "id")
	if got != "" || err == nil {
		t.Error(
			"expected nil",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathNotArrayWithIndex(t *testing.T) {
	s := `{
		"items":
			{"id":"-1","status":"OK"}
	 	}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items", "1", "id")
	if got != "" || err == nil {
		t.Error(
			"expected nil",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathNotIndexWithArray(t *testing.T) {
	s := `{
		"items":[
			{"id":"-1","status":"OK"},
			{"id":"-2","status":"OK"}
		]
	 	}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items", "id")
	if got != "" || err == nil {
		t.Error(
			"expected nil",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathEmpty(t *testing.T) {
	emptyMap := make(map[string]interface{})

	got, _ := getByPath(emptyMap, "token")

	if got != "" {
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
