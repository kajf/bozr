package main

import "testing"

func TestPlainTextWithoutTemplate(t *testing.T) {
	// given
	tmpl := "text-value 123"
	ctx := NewTemplateContext(NewVars())

	// when
	output, err := executeTemplate(ctx, tmpl)

	// then
	expected := "text-value 123"

	if err != nil {
		t.Error("Unexpected error", err.Error())
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

	ctx := NewTemplateContext(vars)

	// when
	output, err := executeTemplate(ctx, tmpl)

	// then
	expected := "Smith was successfully assigned to the Order #555"

	if err != nil {
		t.Error("Unexpected error", err.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncBase64(t *testing.T) {
	// given
	tmpl := "{{ .Base64 `DPFG` }}"

	ctx := NewTemplateContext(NewVars())

	// when
	output, err := executeTemplate(ctx, tmpl)

	// then
	expected := "RFBGRw=="

	if err != nil {
		t.Error("Unexpected error", err.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}

func TestFuncSHA1(t *testing.T) {
	// given
	tmpl := "{{ .SHA1 `{username}` }}"

	vars := NewVars()
	vars.Add("username", "DPFG")

	ctx := NewTemplateContext(vars)

	// when
	output, err := executeTemplate(ctx, tmpl)

	// then
	expected := "f7ae86bd41671ad2e592ed38ab87a043d33e0b84"

	if err != nil {
		t.Error("Unexpected error", err.Error())
	}

	if output != expected {
		t.Errorf("Unexpected output. Expected: %s, Actual: %s", expected, output)
	}
}
