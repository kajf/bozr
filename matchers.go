package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	expectationPathSeparator = "."
	expectationSearchSign    = "~"
)

// GetByPath returns value by exact path line
func GetByPath(m interface{}, pathLine string) (interface{}, error) {

	res := Search(m, pathLine)

	if len(res) != 1 {
		str := fmt.Sprintf("Required exactly one value, found [%v] on path [%v]", len(res), pathLine)
		return nil, errors.New(str)
	}

	if HasPathFunc(pathLine) {
		funcRes, err := CallPathFunc(pathLine, res[0])

		if err == nil {
			return funcRes, nil
		}

		return nil, err
	}

	return res[0], nil
}

// SearchByPath search traversing maps and arrays deep. Returns error with message if expected value not found, nil - otherwise
func SearchByPath(m interface{}, expectedValue interface{}, pathLine string) error {
	//fmt.Println("searchByPath", m, expectedValue, path, reflect.TypeOf(expectedValue))

	resArr := Search(m, pathLine)

	if HasPathFunc(pathLine) {

		if len(resArr) != 1 {
			return fmt.Errorf("Required exactly one result to calculate, found %#v on path %#v", len(resArr), pathLine)
		}

		funcRes, err := CallPathFunc(pathLine, resArr[0])
		if err == nil {
			if funcRes == expectedValue {
				return nil
			}
			return fmt.Errorf("Expected value %#v does not match actual %#v on path %#v", expectedValue, funcRes, pathLine)
		}

		return err
	}

	switch typedExpectedValue := expectedValue.(type) {
	// single path have to match multiple expectations, e.g. items.id : [12,34,56]
	case []interface{}:
		for _, expectedItem := range typedExpectedValue {

			found := findDeep(resArr, expectedItem)

			if !found {
				str := fmt.Sprintf("Value %#v not found on path %#v", expectedItem, pathLine)
				return errors.New(str)
			}
		}

		return nil

	default:
		if findDeep(resArr, expectedValue) {
			return nil
		}
	}

	str := fmt.Sprintf("Value %#v not found on path %#v", expectedValue, pathLine)
	return errors.New(str)
}

func matchesAll(expectedMap map[string]interface{}, searchResult interface{}) bool {

	switch typedSearchRes := searchResult.(type) {
	case map[string]interface{}:

		for field := range expectedMap {
			if expectedMap[field] != typedSearchRes[field] {
				return false
			}
		}

		return true
	}

	return false
}

// Search values represented by (pathLine) recursively at tree object (m)
// returns array of found results (array size = number of results)
// each found result may have be any shape (array, map, value)
func Search(m interface{}, pathLine string) []interface{} {
	path := cleanPath(pathLine)

	res := make([]interface{}, 0)
	search(m, path, &res)

	return res
}

func search(m interface{}, splitPath []string, res *[]interface{}) {
	//fmt.Println(m, "~~~", splitPath)

	if len(splitPath) == 0 {
		*res = append(*res, m)
		return
	} // reached end of path - found

	firstPathPart := splitPath[0]
	if firstPathPart == "" {
		search(m, splitPath[1:], res)
		return
	} // empty path elements do not lead anywhere

	switch typedM := m.(type) {
	case map[string]interface{}:
		if obj, ok := typedM[firstPathPart]; ok {
			search(obj, splitPath[1:], res)
		}

	case []interface{}:
		if idx, err := strconv.Atoi(firstPathPart); err == nil { // index in path
			if idx < len(typedM) { // index exists in array
				search(typedM[idx], splitPath[1:], res)
			}
		} else { // search all items in array
			for _, obj := range typedM {
				search(obj, splitPath, res)
			}
		}

	}
}

func findDeep(items []interface{}, expected interface{}) bool {
	for _, item := range items {

		switch typedItem := item.(type) {
		case []interface{}:
			found := findDeep(typedItem, expected)
			if found {
				return true
			}

		// single path have to match object, e.g. root.items : {id: 1, name: 'example'}
		case map[string]interface{}:
			switch typedExpected := expected.(type) {
			case map[string]interface{}:
				if matchesAll(typedExpected, typedItem) {
					return true
				}
			}

		default:
			if expected == item {
				return true
			}
		}

	}

	return false
}

func cleanPath(pathLine string) []string {
	pathLine = strings.Replace(pathLine, expectationSearchSign, "", -1) // compliance for redundant '~' opeator
	path := strings.Split(pathLine, expectationPathSeparator)

	last := path[len(path)-1]
	lastIsFunc := strings.HasSuffix(last, "()")
	if lastIsFunc {
		path = path[0 : len(path)-1]
	} // remove functions

	return path
}
