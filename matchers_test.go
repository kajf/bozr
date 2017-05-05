package main

import (
	"encoding/json"
	"testing"
)

// --------		path	search	func
// expect		+		+		+
// remember		+		-		+
// absent		+		+		-

func TestPathMapWithIndexes(t *testing.T) {
	m, err := jsonAsMap(
		`{"boo": {
		  "name":"ga-ga"},
		  "items":[
			{
			"id":123,
			"name":"abc"
			}
	      ],
		  "values" : [1,4]
		}`)
	if err != nil {
		t.Error(err)
	}

	res := make(map[string]interface{})
	pathMap("", m, res)

	if res["boo.name"] != "ga-ga" {
		t.Error()
	}

	if res["items.0.id"] != 123.0 {
		t.Error()
	}

	if res["values.1"] != 4.0 {
		t.Error()
	}
}

func TestPathMapWithRootArrayAndIndexes(t *testing.T) {
	arr, err := jsonAsArray(
		`[{"id":123, "name":"abc"},
		  {"id":456, "name":"z"}
	      ]`)
	if err != nil {
		t.Error(err)
	}

	res := make(map[string]interface{})
	pathMap("", arr, res)

	if res["0.name"] != "abc" {
		t.Error()
	}

}

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
	m, err := jsonAsMap(`{
		"root":[{},{}]
	      }`)
	if err != nil {
		t.Error(err)
	}

	found, _ := searchByPath(m, 2.0, "root.size()")

	if !found {
		t.Error()
	}
}

func TestSearchByInvalidPathWithPathFunction(t *testing.T) {
	m, err := jsonAsMap(`{
		"root":[{},{}]
	}`)
	if err != nil {
		t.Error(err)
	}

	found, err := searchByPath(m, nil, "notExist.size()")

	if found || err == nil {
		t.Error(
			"expected found=false + path error",
			"got", found,
			"err", err,
		)
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

func TestSearchByPathWithRootArray(t *testing.T) {
	arr, err := jsonAsArray(`[
			{"id":1,"status":"OK"},
			{"id":2,"status":"OK"}
		]`)
	if err != nil {
		t.Error(err)
	}

	expected := []interface{}{2.0}
	ok, err := searchByPath(arr, expected, "id")
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

func TestGetByInvalidPathFunction(t *testing.T) {
	m, err := jsonAsMap(`{
		"items":[{},{}]
	 	}`)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m, "notExist.size()")
	if got != nil || err == nil {
		t.Error(
			"expected nil",
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

func TestGetByPathWithArray(t *testing.T) {
	arr, err := jsonAsArray(`[
			{"id":"abc","status":"OK"},
			{"id":"zz","status":"OK"}
		]`)
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(arr, "1.id")
	if got != "zz" || err != nil {
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

func jsonAsArray(s string) ([]interface{}, error) {
	arr := make([]interface{}, 0)
	err := json.Unmarshal([]byte(s), &arr)

	return arr, err
}

func jsonAsMap(s string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(s), &m)

	return m, err
}
