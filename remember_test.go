package main

import "testing"

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
