package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	pathFuncs = map[string]pathFunc{
		"size()":         size,
		"string()":       str,
		"sizeAsString()": sizeAsStr,
	}
)

// HasPathFunc checks whether path contains a function or not
func HasPathFunc(pathLine string) bool {

	for fname := range pathFuncs {
		if !strings.HasSuffix(pathLine, fname) {
			continue
		}

		return true
	}

	return false
}

// CallPathFunc executes suffix function from pathLine passing given arg
func CallPathFunc(pathLine string, arg interface{}) (interface{}, error) {

	for fname, f := range pathFuncs {
		if !strings.HasSuffix(pathLine, fname) {
			continue
		}

		return f(arg)
	}

	return nil, fmt.Errorf("No function declarations found on path %#v", pathLine)
}

type pathFunc func(arg interface{}) (interface{}, error)

func size(arg interface{}) (interface{}, error) {

	switch arr := arg.(type) {
	case []interface{}:
		return float64(len(arr)), nil

	default:
		str := fmt.Sprintf("size() is not applicable to arg %#v", arg)
		return nil, errors.New(str)
	}
}

func str(arg interface{}) (interface{}, error) {
	return toString(arg), nil
}

func sizeAsStr(arg interface{}) (interface{}, error) {
	numSize, err := size(arg)
	if err != nil {
		return nil, err
	}

	return toString(numSize), nil
}
