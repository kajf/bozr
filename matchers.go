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
	pathSizeFunction         = "size()"
)

// GetByPath returns value by exact path line
func GetByPath(m interface{}, pathLine string) (interface{}, error) {

	res := Search(m, pathLine)

	if len(res) != 1 {
		str := fmt.Sprintf("Required exactly one value, found [%v] on path [%v]", len(res), pathLine)
		return nil, errors.New(str)
	}

	if strings.HasSuffix(pathLine, pathSizeFunction) {
		currSize, err := calcSize(pathLine, res)

		if err == nil {
			return currSize, nil
		}

		return nil, err
	}

	return res[0], nil
}

// SearchByPath search traversing maps and arrays deep
func SearchByPath(m interface{}, expectedValue interface{}, pathLine string) (bool, error) {
	//fmt.Println("searchByPath", m, expectedValue, path, reflect.TypeOf(expectedValue))

	res := Search(m, pathLine)

	if strings.HasSuffix(pathLine, pathSizeFunction) {
		currSize, err := calcSize(pathLine, res)

		if err == nil {
			if currSize == expectedValue {
				return true, nil
			}

			str := fmt.Sprintf("expected [%v].size() [%v] does not match actual [%v]", pathLine, expectedValue, currSize)
			return false, errors.New(str)
		}

		return false, err
	}

	switch typedExpectedValue := expectedValue.(type) {
	case []interface{}:
		//found := false
		for _, expectedItem := range typedExpectedValue {

			found := findDeep(res, expectedItem)

			if !found {
				str := fmt.Sprintf("Value [%v] not found by path [%v]", expectedItem, pathLine)
				return false, errors.New(str)
			}
		}

		return true, nil
	default:
		if findDeep(res, expectedValue) {
			return true, nil
		}
	}

	str := fmt.Sprintf("Value [%v] not found by path [%v]", expectedValue, pathLine)
	return false, errors.New(str)
}

func calcSize(pathLine string, res []interface{}) (float64, error) {

	if !strings.HasSuffix(pathLine, pathSizeFunction) {
		str := fmt.Sprintf("Path has no size function [%v] to calculate", pathLine)
		return -1.0, errors.New(str)
	}

	if len(res) != 1 {
		str := fmt.Sprintf("Required exactly one value to calculate, found [%v] on path [%v]", len(res), pathLine)
		return -2.0, errors.New(str)
	}

	switch arr := res[0].(type) {
	case []interface{}:
		return float64(len(arr)), nil

	default:
		str := fmt.Sprintf(".size() is not applicable to search result [%v] ", res)
		return -3.0, errors.New(str)
	}
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
	pathLine = strings.TrimSuffix(pathLine, pathSizeFunction)           // function

	path := strings.Split(pathLine, expectationPathSeparator)

	return path
}
