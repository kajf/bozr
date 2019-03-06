package main

import (
	"encoding/json"
	"testing"
)

// --------		path	search	func
// expect		+		+		+
// remember		+		-		+
// absent		+		+		-

func TestSearch(t *testing.T) {
	m, err := jsonAsMap(
		`{
			"boo": {"name":"ga-ga"}
		 }`)
	if err != nil {
		t.Error(err)
	}

	res := Search(m, "boo.name")

	if len(res) != 1 || res[0] != "ga-ga" {
		t.Error("unexpected ", res)
	}
}

func TestSearchMultiResult(t *testing.T) {
	m, err := jsonAsMap(
		`{"boo": { "name":"ga-ga"},
		  "items":[
			{"id":123, "name":"abc"},
			{"id":45, "name":"de"}
	      ]
		}`)
	if err != nil {
		t.Error(err)
	}

	res := Search(m, "items.id")

	if len(res) != 2 || res[0] != 123.0 {
		t.Error("unexpected ", res)
	}
}

func TestSearchWithIndex(t *testing.T) {
	m, err := jsonAsMap(
		`{
		  "items":[
			{"id":123, "name":"abc"},
			{"id":45, "name":"de"}
	      ]
		}`)
	if err != nil {
		t.Error(err)
	}

	res := Search(m, "items.1.id")

	if len(res) != 1 || res[0] != 45.0 {
		t.Error("unexpected ", res)
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

	err = SearchByPath(m, 417601.0, "rate_tables.id")

	if err != nil {
		t.Error(err)
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

	err = SearchByPath(m, "-1", "root.key")

	if err != nil {
		t.Error(err)
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

	err = SearchByPath(m, "test 2", "root.name")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathArray(t *testing.T) {
	m, err := jsonAsMap(`{
		"root":[{},{}]
	      }`)
	if err != nil {
		t.Error(err)
	}

	err = SearchByPath(m, 2.0, "root.size()")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchBySizeEmptyRootArray(t *testing.T) {
	m, err := jsonAsArray(`[{},{}]`)
	if err != nil {
		t.Error(err)
	}

	err = SearchByPath(m, 2.0, "size()")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathDuplicates(t *testing.T) {
	m, err := jsonAsMap(`{
					"counters": [
						{
							"counters": [{},{},{}]
						}
					]
			}`)
	if err != nil {
		t.Error(err)
	}

	err = SearchByPath(m, 3.0, "counters.counters.size()")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathWithObjectFieldsFromDifferentItems(t *testing.T) {
	m, err := jsonAsMap(`{
					"items": [
						{
							"id": 12,
							"name": "foo",
							"descr": "abc"
						},
						{
							"id": 34,
							"name": "bar",
							"descr": "bbb"
						}						
					]
			}`)
	if err != nil {
		t.Error(err)
	}

	expect := map[string]interface{}{
		"id":   12.0,
		"name": "bar",
	}

	err = SearchByPath(m, expect, "items")

	if err == nil {
		t.Error("unexpected foundation. name and id have to be from the same item")
	}
}

func TestSearchByPathWithObject(t *testing.T) {
	m, err := jsonAsMap(`{
					"items": [
						{
							"id": 12,
							"name": "foo",
							"descr": "abc"
						},
						{
							"id": 34,
							"name": "bar",
							"descr": "bbb"
						}						
					]
			}`)
	if err != nil {
		t.Error(err)
	}

	expect := map[string]interface{}{
		"id":   34.0,
		"name": "bar",
	}

	err = SearchByPath(m, expect, "items")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathWithObjectByIndex(t *testing.T) {
	m, err := jsonAsMap(`{
					"items": [
						{
							"id": 12,
							"name": "foo",
							"descr": "abc"
						},
						{
							"id": 34,
							"name": "bar",
							"descr": "bbb"
						}						
					]
			}`)
	if err != nil {
		t.Error(err)
	}

	expect := map[string]interface{}{
		"id":   34.0,
		"name": "bar",
	}

	err = SearchByPath(m, expect, "items.1")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathWithArrayOfObjects(t *testing.T) {
	m, err := jsonAsMap(`{
					"items": [
						{
							"id": 12,
							"name": "foo"
						},
						{
							"id": 34,
							"name": "bar"
						},
						{
							"id": 56,
							"name": "baz"
						}
											
					]
			}`)
	if err != nil {
		t.Error(err)
	}

	expect := []interface{}{

		map[string]interface{}{
			"id":   34.0,
			"name": "bar",
		},

		map[string]interface{}{
			"id":   12.0,
			"name": "foo",
		},
	}

	err = SearchByPath(m, expect, "items")

	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathWithMultiArraysObject(t *testing.T) {
	m, err := jsonAsMap(`{
				"lookups": [
					{
						"name": "first",
						"items": [
							{
								"id": 12,
								"name": "foo",
								"descr": "abc"
							},
							{
								"id": 34,
								"name": "bar",
								"descr": "bbb"
							}						
						]
					},
					{
						"name": "second",
						"items": [
							{
								"id": 56,
								"name": "baz"
							}						
						]
					}
				]
			}`)
	if err != nil {
		t.Error(err)
	}

	expect := map[string]interface{}{
		"id":   56.0,
		"name": "baz",
	}

	err = SearchByPath(m, expect, "lookups.items")

	if err != nil {
		t.Error(err)
	}
}

func TestMatchesAllNonMap(t *testing.T) {

	expect := map[string]interface{}{
		"id":   34,
		"name": "bar",
	}

	expectMore := map[string]interface{}{
		"id":    34,
		"name":  "bar",
		"descr": "hey",
	}

	var flagtests = []struct {
		in  interface{}
		out bool
	}{
		{"", false},
		{"ab", false},
		{12, false},
		{nil, false},
		{[]interface{}{}, false},
		{[]interface{}{34, 2}, false},
		{expect, true},
		{expectMore, true},
	}

	for _, tt := range flagtests {
		t.Run("matchesAll", func(t *testing.T) {

			matches := matchesAll(expect, tt.in)

			if matches != tt.out {
				t.Errorf("matchAll(&expect, %v) => %v, want %v",
					tt.in, matches, tt.out)
			}
		})
	}
}

func TestSearchByInvalidPathWithPathFunction(t *testing.T) {
	m, err := jsonAsMap(`{
		"root":[{},{}]
	}`)
	if err != nil {
		t.Error(err)
	}

	err = SearchByPath(m, nil, "notExist.size()")

	if err == nil {
		t.Error(
			"expected foundation", err,
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

	err = SearchByPath(m, "-2", "second.key")

	if err != nil {
		t.Error(err)
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

	err = SearchByPath(m, "-2", "single.key")

	if err == nil {
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
	err = SearchByPath(m, arr, "items.id")
	if err != nil {
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
	err = SearchByPath(m, arr, "items.id")
	if err == nil {
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
	err = SearchByPath(m, arr, "items.id")
	if err != nil {
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
	err = SearchByPath(m, arr, "items.id")
	if err != nil {
		t.Error(err)
	}
}

func TestFindDeep(t *testing.T) {
	sub := []interface{}{"def"}
	root := []interface{}{sub}

	found := findDeep(root, "def")

	if !found {
		t.Error()
	}
}

func TestFindDeepInFlat(t *testing.T) {
	root := []interface{}{"def", "ab"}

	found := findDeep(root, "ab")

	if !found {
		t.Error()
	}
}

func TestSearchByPathArrayOfPrimitives(t *testing.T) {
	m, err := jsonAsMap(`{"items":["ONE", "TWO"]}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{"ONE", "TWO"}
	err = SearchByPath(m, arr, "items")
	if err != nil {
		t.Error(err)
	}
}

func TestSearchByPathArrayOfPrimitivesSingle(t *testing.T) {
	m, err := jsonAsMap(`{"items":["ONE", "TWO"]}`)
	if err != nil {
		t.Error(err)
	}

	arr := []interface{}{"ONE"}
	err = SearchByPath(m, arr, "items")
	if err != nil {
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
	err = SearchByPath(m, arr, "items.id")
	if err != nil {
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
	err = SearchByPath(arr, expected, "id")
	if err != nil {
		t.Error(err)
	}
}

func TestGetByPathSimple(t *testing.T) {
	token := "abc"

	m, err := jsonAsMap(`{"token":"` + token + `","ttl":3600000,"units":"milliseconds"}`)
	if err != nil {
		t.Error(err)
	}

	got, _ := GetByPath(m, "token")

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

	got, _ := GetByPath(m, "token.name")

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

	got, _ := GetByPath(m, "items.0.id")
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

	got, err := GetByPath(m, "items.size()")
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

	got, err := GetByPath(m, "notExist.size()")
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

	got, err := GetByPath(m, "items.2.id")
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

	got, err := GetByPath(m, "items.1.id")
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

	got, err := GetByPath(m, "items.id")
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

	got, err := GetByPath(arr, "1.id")
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

	got, _ := GetByPath(emptyMap, "token")

	if got != nil {
		t.Error(
			"For", "token",
			"expected", nil,
			"got", got,
		)
	}
}

func TestGetByPathEmptyRootArraySize(t *testing.T) {
	m, err := jsonAsMap(`{
				"items":[]
			}`)

	if err != nil {
		t.Error(err)
	}

	got, _ := GetByPath(m, "items.size()")

	if got != 0.0 {
		t.Error(
			"For items.size",
			"expected", 0.0,
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

	_, err = GetByPath(m, "rates.z")
	if err == nil {
		t.Error(
			"For", "rates.z",
			"expected", "error",
			"got", err,
		)
	}
}

func TestCleanPathFuncTrimmed(t *testing.T) {
	path := cleanPath("some.ids.myFunc()")

	last := path[len(path)-1]
	if last == "myFunc()" {
		t.Error("Function should not be the last item in path", path)
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

func _jsonAsMap(s string) map[string]interface{} {
	v, _ := jsonAsMap(s)
	return v
}

func TestBodyMatch(t *testing.T) {

	data := []struct {
		name         string
		body         interface{}
		matcher      interface{}
		strict       bool
		expectToFail bool
	}{
		{
			name:    "Object/Partial/Matches",
			strict:  false,
			body:    _jsonAsMap(`{ "profile": { "name": "Jack", "age": 31 } }`),
			matcher: _jsonAsMap(`{ "profile": { "name": "Jack" } }`),
		},
		{
			name:         "Object/Partial/FailIfAtLeastOneDoesntMatch",
			strict:       false,
			body:         _jsonAsMap(`{ "profile": { "name": "Jack", "age": 31 } }`),
			matcher:      _jsonAsMap(`{ "profile": { "name": "John" } }`),
			expectToFail: true,
		},
		{
			name:         "Object/Partial/FailIfAtLeastOneIsMissing",
			strict:       false,
			body:         _jsonAsMap(`{ "profile": { "name": "Jack", "age": 31 } }`),
			matcher:      _jsonAsMap(`{ "profile": { "name": "Jack", "sex": "male" } }`),
			expectToFail: true,
		},
		{
			name:         "Object/Exact/FailIfAtLeastOneDoesntMatch",
			strict:       true,
			body:         _jsonAsMap(`{ "profile": { "name": "Jack", "age": 31 } }`),
			matcher:      _jsonAsMap(`{ "profile": { "name": "John", "age": 31 } }`),
			expectToFail: true,
		},
		{
			name:         "Object/Exact/FailIfAtLeastOneIsMissing",
			strict:       true,
			body:         _jsonAsMap(`{ "profile": { "name": "Jack", "age": 31 } }`),
			matcher:      _jsonAsMap(`{ "profile": { "name": "Jack" } }`),
			expectToFail: true,
		},

		{
			name:         "Array/Partial/IntegersMatchesInAnyOrder",
			strict:       false,
			body:         _jsonAsMap(`{ "items": [1,2,3,4,5] }`),
			matcher:      _jsonAsMap(`{ "items": [2,1,3,5] }`),
			expectToFail: false,
		},

		{
			name:         "Array/Exact/IntegersMatchesIfOrdered",
			strict:       true,
			body:         _jsonAsMap(`{ "items": [1,2,3,4,5] }`),
			matcher:      _jsonAsMap(`{ "items": [1,2,3,4,5] }`),
			expectToFail: false,
		},
		{
			name:         "Array/Exact/IntegersFailsIfAtLeastOnIsMissing",
			strict:       true,
			body:         _jsonAsMap(`{ "items": [1,2,3,4,5] }`),
			matcher:      _jsonAsMap(`{ "items": [1,2,3,4] }`),
			expectToFail: true,
		},
		{
			name:         "Array/Exact/IntegersFailsIfAtLeastOneIsOutOfOrder",
			strict:       true,
			body:         _jsonAsMap(`{ "items": [1,2,3,4,5] }`),
			matcher:      _jsonAsMap(`{ "items": [1,2,4,3,5] }`),
			expectToFail: true,
		},
	}

	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			expectation := NewBodyMatcher{ExpectedBody: tt.matcher, Strict: tt.strict}
			err := expectation.check(tt.body)

			if tt.expectToFail && err == nil {
				t.Error("Expected to dont match")
			}

			if !tt.expectToFail && err != nil {
				t.Error("Expected to match")
			}
		})
	}
}
