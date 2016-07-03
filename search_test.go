package main

import (
	"testing"
	"encoding/json"
)

func TestSearchByPathId(t *testing.T) {

	s := `{"rate_tables":[
		{
		"id":417601,
		"name":"Test Rate Table1"
		}
	      ]}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	found := searchByPath(m, "417601", "rate_tables", "id")

	if !found {
		t.Error()
	}
}

func TestSearchByPathKey(t *testing.T) {

	s := `{"root":[
		{
		"key":"-1",
		"name":"Test"
		}
	      ]}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	found := searchByPath(m, "-1", "root", "key")

	if !found {
		t.Error()
	}
}

func TestSearchByPathArray(t *testing.T) {

	s := `{"root":[
		{"key":"-1", "name":"Test 1"},
		{"key":"-2", "name":"test 2"}
	      ]}`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	found := searchByPath(m, "test 2", "root", "name")

	if !found {
		t.Error()
	}
}

func TestSearchByPathSingleObject(t *testing.T) {

	s := `{
		"first":{
			"key":"-1",
			"name":"Test 1"
			},
		"second" : {
			"key":"-2",
			"name":"test 2"
			}
	      }`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	found := searchByPath(m, "-2", "second", "key")

	if !found {
		t.Error()
	}
}

func TestSearchByPathNotFound(t *testing.T) {

	s := `{
		"single":{
			"key":"-1",
			"name":"Test 1"
			}
	      }`
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	found := searchByPath(m, "-2", "single", "key")

	if found {
		t.Error()
	}
}