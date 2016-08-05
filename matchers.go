package main

import (
	"fmt"
	"strconv"
	"github.com/pkg/errors"
	"strings"
)

type MatcherFunc func(root interface{}, expectedValue interface{}, path  ...string) (bool, error)

func ChooseMatcher(path string) MatcherFunc {
	exactMatch := !strings.HasPrefix(path, "~")

	if exactMatch {
		return equalsByPath
	} else {
		return searchByPath
	}
}

func equalsByPath(m interface{}, expectedValue interface{}, path ...string) (bool, error) {

	switch typedExpectedValue := expectedValue.(type) {
	case string:
		val, err := getByPath(m, path...)
		return (typedExpectedValue == val), err
	}

	return false, nil
}

// exact value by exact path
func getByPath(m interface{}, path ...string) (string, error) {

	for _, p := range path {
		//fmt.Println(p)
		funcVal, ok := pathFunction(m, p)
		if ok {
			return funcVal, nil
		}

		idx, err := strconv.Atoi(p)
		if err != nil {
			//fmt.Println(err)
			mp, ok := m.(map[string]interface{})
			if !ok {
				str := fmt.Sprintf("Can't cast to Map and get key [%v] in path %v", p, path)
				return "", errors.New(str)
			}
			m = mp[p]
		} else {
			arr, ok := m.([]interface{})
			if !ok {
				str := fmt.Sprintf("Can't cast to Array and get index [%v] in path %v", idx, path)
				return "", errors.New(str)
			}
			if idx >= len(arr) {
				str := fmt.Sprintf("Array only has [%v] elements. Can't get element by index [%v] (counts from zero)", len(arr), idx)
				return "", errors.New(str)
			}
			m = arr[idx]
		}
	}

	if str, ok := castToString(m); ok {
		return str, nil
	}
	strErr := fmt.Sprintf("Can't cast path result to string: %v", m)
	return "", errors.New(strErr)
}

// search passing maps and arrays
func searchByPath(m interface{}, expectedValue interface{}, path ...string) (bool, error) {
	//fmt.Println("[",expectedValue, "] ", reflect.TypeOf(expectedValue))
	switch typedExpectedValue := expectedValue.(type) {
	case []interface{}:
		for _, obj := range typedExpectedValue {
			if ok, err := searchByPath(m, obj, path...); !ok {
				return false, err
			}
		}
		return true, nil
	case string:
		for idx, p := range path {
			//fmt.Println("s ", idx, "p ", p)
			funcVal, ok := pathFunction(m, p)
			if ok {
				if typedExpectedValue == funcVal {
					return true, nil
				}
			}

			switch typedM := m.(type) {
			case map[string]interface{}:
				m = typedM[p]
				//fmt.Println("[",m, "] ", reflect.TypeOf(m))
				if str, ok := castToString(m); ok {
					if str == typedExpectedValue {
						return true, nil
					}
				}
			case []interface{}:
				//fmt.Println("path ", path[idx:])
				for _, obj := range typedM {
					found, err := searchByPath(obj, typedExpectedValue, path[idx:]...)
					if found {
						return true, err
					}
				}
			}
		}
	}
	return false, nil
}

func castToString(m interface{}) (string, bool) {
	//fmt.Println("[",m, "] ", reflect.TypeOf(m))
	if str, ok := m.(string); ok {
		return str, ok
	} else if flt, ok := m.(float64); ok {
		// numbers (like ids) are parsed as float64 from json
		return strconv.FormatFloat(flt, 'f', 0, 64), ok
	} else {
		return "", ok
	}
}

func pathFunction(m interface{}, pathPart string) (string, bool) {

	if pathPart == "size()" {
		if arr, ok := m.([]interface{}); ok {
			return strconv.Itoa(len(arr)), true
		}
	}

	return "", false
}