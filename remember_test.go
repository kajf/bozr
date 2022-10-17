package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPopulateRequestBody(t *testing.T) {
	//given
	on := On{URL: "http://example.com"}
	value := "abc"
	vars := NewVars("")
	vars.AddAll(map[string]interface{}{"var": value})

	tmplCtx := NewTemplateContext(vars)
	body := tmplCtx.ApplyTo("pre {var} post")

	// when
	req, _ := populateRequest(&RequestConfig{}, on, body, tmplCtx)

	//then
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	got := buf.String()
	if !strings.Contains(got, value) {
		t.Error(
			"body does not conatain value:", value,
			"got:", got,
		)
	}
}

func TestConvertTypesToString(t *testing.T) {
	makeTest := func(val interface{}, expected string) func(t *testing.T) {
		return func(t *testing.T) {
			got := toString(val)

			if got != expected {
				t.Error(
					"expected[", expected, "]",
					"got[", got, "]",
				)
			}
		}
	}

	t.Run("int", makeTest(1, "1"))
	t.Run("float", makeTest(1.00001, "1.00001"))
	t.Run("boolean", makeTest(false, "false"))
	t.Run("string", makeTest("example", "example"))
}

func TestRememberHeader(t *testing.T) {
	responseHeaders := map[string][]string{"X-Test": {"PASS"}}
	remember := map[string]string{"valueKey": "X-Test"}

	vars := NewVars("")
	initialLen := len(vars.items)

	rememberHeaders(responseHeaders, remember, vars)

	if len(vars.items)-initialLen != 1 {
		t.Errorf("Unexpected map length: %d", len(remember))
		return
	}

	if vars.items["valueKey"] != "PASS" {
		t.Errorf("Unexpected remembered value: %s", vars.items["valueKey"])
	}
}
