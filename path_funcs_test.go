package main

import "testing"

type hasPathFuncTest struct {
	path     string
	expected bool
}

var hasPathFuncTests = []hasPathFuncTest{
	{"size()", true},
	{"my.path.string()", true},
	{"some.path", false},
	{"some.path", false},
	{"items.1.path", false},
	{"items.1.path.size()", true},
	{"items.path.xyz()", false},
}

func TestHasPathFunc(t *testing.T) {

	for _, tt := range hasPathFuncTests {

		actual := HasPathFunc(tt.path)

		if actual != tt.expected {
			t.Errorf("HasPathFunc(%#v): expected %#v, actual %#v", tt.path, tt.expected, actual)
		}
	}

}

func TestCallPathFuncStr(t *testing.T) {

	pathLine := "items.id.string()"
	arg := 123.0
	res, err := CallPathFunc(pathLine, arg)

	expected := "123"
	if res != expected || err != nil {
		t.Errorf("Expected %s Got %#v, %#v = CallPathFunc(%#v, %#v)", expected, res, err, pathLine, arg)
	}
}

func TestCallPathFuncStrArr(t *testing.T) {

	pathLine := "items.string()"
	arg := []string{"1", "2"}

	res, err := CallPathFunc(pathLine, arg)

	expected := "[1 2]"
	if res != expected || err != nil {
		t.Errorf("Expected %s Got %#v, %#v = CallPathFunc(%#v, %#v)", expected, res, err, pathLine, arg)
	}
}

func TestCallPathFuncSizeArr(t *testing.T) {

	pathLine := "items.size()"
	arg := []interface{}{5, 8}

	res, err := CallPathFunc(pathLine, arg)

	var expected float64 = 2 // float (not int) for json parser always return it for numbers
	if res != expected || err != nil {
		t.Errorf("Expected %#v Got %#v, %#v = CallPathFunc(%#v, %#v)", expected, res, err, pathLine, arg)
	}
}
