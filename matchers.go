package main

import (
	"fmt"
	"strconv"
	"github.com/pkg/errors"
	"strings"
)

type MatcherFunc func(root interface{}, expectedValue string, path  ...string) (bool, error)

func ChooseMatcher(path string) MatcherFunc {
	exactMatch := !strings.HasPrefix(path, "~")

	if exactMatch {
		return equalsByPath
	} else {
		return searchByPath
	}
}

func equalsByPath(m interface{}, expectedValue string, path ...string) (bool, error) {
	val, err := getByPath(m, path...)
	return (expectedValue == val), err
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
func searchByPath(m interface{}, s string, path ...string) (bool, error) {
	for idx, p := range path {
		//fmt.Println("s ", idx, "p ", p)
		funcVal, ok := pathFunction(m, p)
		if ok {
			if s == funcVal {
				return true, nil
			}
		}

		switch typedM := m.(type) {
		case map[string]interface{}:
			m = typedM[p]
			//fmt.Println("[",m, "] [", s,"]", reflect.TypeOf(m))
			if str, ok := castToString(m); ok {
				if str == s {
					return true, nil
				}
			}
		case []interface{}:
			//fmt.Println("path ", path[idx:])
			for _, obj := range typedM {
				found, err := searchByPath(obj, s, path[idx:]...)
				if found {
					return true, err
				}
			}
		}
	}

	return false, nil
}