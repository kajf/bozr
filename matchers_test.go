package main

import (
	"encoding/json"
	"testing"
)

func TestSearchByPathId(t *testing.T) {

	m, err := jsonAsMap(`{"rate_tables":[
		{
		"id":417601,
		"name":"Test Rate Table1"
		}
	      ]}`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, 417601.0, "rate_tables.id")

	if !found {
		t.Error()
	}
}

func TestSearchByPathKey(t *testing.T) {

	m, err := jsonAsMap(
		`{"root":[
				{
				"key":"-1",
				"name":"Test"
				}
	      ]}`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, "-1", "root.key")

	if !found {
		t.Error()
	}
}

func TestSearchByPathInArray(t *testing.T) {

	m, err := jsonAsMap(`{"root":[
		{"key":"-1", "name":"Test 1"},
		{"key":"-2", "name":"test 2"}
	      ]}`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, "test 2", "root.name")

	if !found {
		t.Error()
	}
}

func TestSearchByPathArray(t *testing.T) {
	m, err := jsonAsMap(`{"root":[
		{"key":"-1", "name":"Test 1"},
		{"key":"-2", "name":"test 2"}
	      ]}`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, 2.0, "root.size()")

	if !found {
		t.Error()
	}
}

func TestSearchByPathSingleObject(t *testing.T) {

	m, err := jsonAsMap(`{
		"first":{
			"key":"-1",
			"name":"Test 1"
			},
		"second" : {
			"key":"-2",
			"name":"test 2"
			}
	  }`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, "-2", "second.key")

	if !found {
		t.Error()
	}
}

func TestSearchByPathNotFound(t *testing.T) {

	m, err := jsonAsMap(`{
		"single":{
			"key":"-1",
			"name":"Test 1"
			}
	    }`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, "-2", "single.key")

	if found {
		t.Error()
	}
}

func TestSearchByPathExactHasArray(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"a","status":"OK"},
			{"id":"b","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{"a", "b"}
	ok, err := searchByPath(m, arr, "items.id")
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestSearchByPathHasNotAllArrayItems(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"a","status":"OK"},
			{"id":"b","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	arr := []string{"a", "b", "c"}
	ok, err := searchByPath(m, arr, "items.id")
	if ok {
		t.Error("Should have failed because of 'c'")
	}
}

func TestSearchByPathInLargerSet(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"a","status":"OK"},
			{"id":"b","status":"OK"},
			{"id":"c","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{"a", "b"}
	ok, err := searchByPath(m, arr, "items.id")
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestSearchByPathHasOneElementArray(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"a","status":"OK"},
			{"id":"b","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{"a"}
	ok, err := searchByPath(m, arr, "items.id")
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestSearchByPathArrayOfPrimitives(t *testing.T) {
	m, err := jsonAsMap(`{"items":["ONE", "TWO"]}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{"ONE", "TWO"}
	ok, err := searchByPath(m, arr, "items")
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestSearchByPathHasIntArr(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":1,"status":"OK"},
			{"id":2,"status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{1.0, 2.0}
	ok, err := searchByPath(m, arr, "items.id")
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestGetByPathSimple(t *testing.T) {
	token := "abc"

	m, err := jsonAsMap(`{"token":"` + token + `","ttl":3600000,"units":"milliseconds"}`)
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

	got, _ := getByPath(m, "token.name")

	if got != name {
		t.Error(
			"For", "token.name",
			"expected", name,
			"got", got,
		)
	}
}

func TestGetByPathWithIndex(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"417857","status":"OK"},
			{"id":"417858","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	got, _ := getByPath(m, "items.0.id")
	if got != "417857" {
		t.Error(
			"expected", "417857",
			"got", got,
		)
	}
}

func TestGetByPathArraySize(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"status":"OK"},
			{"status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items.size()")
	if got != 2.0 || err != nil {
		t.Error(
			"expected 2",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathArrayOutOfBounds(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"-1","status":"OK"},
			{"id":"-2","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items.2.id")
	if got != nil || err == nil {
		t.Error(
			"expected nil",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathNotArrayWithIndex(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":
			{"id":"-1","status":"OK"}
	 	}`)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items.1.id")
	if got != nil || err == nil {
		t.Error(
			"expected nil",
			"got", got,
			"err", err,
		)
	}
}

func TestGetByPathNotIndexWithArray(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[
			{"id":"-1","status":"OK"},
			{"id":"-2","status":"OK"}
		]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "items.id")
	if got != nil || err == nil {
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

	if got != nil {
		t.Error(
			"For", "token",
			"expected", nil,
			"got", got,
		)
	}
}

func TestGetByPathWithPartialMatch(t *testing.T) {
	m, err := jsonAsMap(`{
				"rates":{
					"AUD":1.4406,
					"BGN":1.9558
				}
			}`)

	if err != nil {
		t.Error(err)
	}

	_, err = getByPath(m, "rates.z")
	if err == nil {
		t.Error(
			"For", "rates.z",
			"expected", "error",
			"got", err,
		)
	}
}

func jsonAsMap(s string) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)

	return m, err
}
