package main

import (
	"fmt"
	"testing"
)

func TestPlainTextWithoutTemplate(t *testing.T) {
	// given
	tmpl := "text-value 123"
	tmplCtx := NewTemplateContext(NewVars())

	// when
	output := tmplCtx.ApplyTo(tmpl)

	// then
	expected := "text-value 123"

	if tmplCtx.HasErrors() {
		t.Error("Unexpected error", tmplCtx.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestPlainTextWithVars(t *testing.T) {
	// given
	tmpl := "{username} was successfully assigned to the Order #{order-id}"

	vars := NewVars()
	vars.Add("order-id", 555)
	vars.Add("username", "Smith")

	tmplCtx := NewTemplateContext(vars)

	// when
	output := tmplCtx.ApplyTo(tmpl)

	// then
	expected := "Smith was successfully assigned to the Order #555"

	if tmplCtx.HasErrors() {
		t.Error("Unexpected error", tmplCtx.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestPlainTextWithNotExistingFunc(t *testing.T) {
	// given
	tmpl := "{username} was successfully assigned to the Order {{ .NotExists `9302945` }}"

	vars := NewVars()
	vars.Add("username", "Smith")

	tmplCtx := NewTemplateContext(vars)

	// when
	output := tmplCtx.ApplyTo(tmpl)

	// then
	expected := ""

	if !tmplCtx.HasErrors() {
		t.Error("Expected error not found")
	}

	if output != "" {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncBase64(t *testing.T) {
	// given
	tmpl := "{{ .Base64 `DPFG` }}"

	tmplCtx := NewTemplateContext(NewVars())

	// when
	output := tmplCtx.ApplyTo(tmpl)

	// then
	expected := "RFBGRw=="

	if tmplCtx.HasErrors() {
		t.Error("Unexpected error", tmplCtx.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncSHA1(t *testing.T) {
	// given
	tmpl := "{{ .SHA1 `{username}` }}"

	vars := NewVars()
	vars.Add("username", "el_mask")

	tmplCtx := NewTemplateContext(vars)

	// when
	output := tmplCtx.ApplyTo(tmpl)

	// then
	expected := "2b0cc371b76f3ec6c1bebc52bcc44af69304dabf"

	if tmplCtx.HasErrors() {
		t.Error("Unexpected error", tmplCtx.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncWSSEPasswordDigest(t *testing.T) {
	// given
	tmpl := "{{ .WSSEPasswordDigest `{nonce}` `{created}` `{password}` }}"

	vars := NewVars()
	vars.Add("nonce", "abc123")
	vars.Add("created", "2012-06-09T18:41:03.640Z")
	vars.Add("password", "password")

	tmplCtx := NewTemplateContext(vars)

	// when
	output := tmplCtx.ApplyTo(tmpl)

	// then
	expected := "mh7Ix8Qe02z1FIr51zoRO5pDMJg="

	if tmplCtx.HasErrors() {
		t.Error("Unexpected error", tmplCtx.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncDaysFromNow(t *testing.T) {
	// given
	const givenTemplate = `increment {{-3 | .DaysFromNow | .FormatDateTime "2006-01-02" }}`

	vars := NewVars()
	tmplCtx := NewTemplateContext(vars)
	funcs := NewFuncs(vars)

	expected := funcs.FormatDateTime("2006-01-02", funcs.DaysFromNow(-3))

	// when
	output := tmplCtx.ApplyTo(givenTemplate)

	if output != fmt.Sprintf("increment %s", expected) {
		t.Error(output, "is not equal to", expected)
	}
}
