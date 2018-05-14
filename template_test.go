package main

import "testing"

func TestPlainTextWithoutTemplate(t *testing.T) {
	// given
	tmpl := "text-value 123"
	proc := NewTemplateProcessor(NewVars())

	// when
	output := proc.ApplyTo(tmpl)

	// then
	expected := "text-value 123"

	if proc.HasErrors() {
		t.Error("Unexpected error", proc.Error())
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

	proc := NewTemplateProcessor(vars)

	// when
	output := proc.ApplyTo(tmpl)

	// then
	expected := "Smith was successfully assigned to the Order #555"

	if proc.HasErrors() {
		t.Error("Unexpected error", proc.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncBase64(t *testing.T) {
	// given
	tmpl := "{{ .Base64 `DPFG` }}"

	proc := NewTemplateProcessor(NewVars())

	// when
	output := proc.ApplyTo(tmpl)

	// then
	expected := "RFBGRw=="

	if proc.HasErrors() {
		t.Error("Unexpected error", proc.Error())
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

	proc := NewTemplateProcessor(vars)

	// when
	output := proc.ApplyTo(tmpl)

	// then
	expected := "2b0cc371b76f3ec6c1bebc52bcc44af69304dabf"

	if proc.HasErrors() {
		t.Error("Unexpected error", proc.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}
