package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	"github.com/pkg/errors"
)

type TemplateContext struct {
	vars *Vars
}

// NewTemplateContext creates new TemplateContext struct that includes initialized Vars.
func NewTemplateContext(vars *Vars) *TemplateContext {
	return &TemplateContext{vars: vars}
}

//
func (ctx *TemplateContext) CurrentTimestampMS() int64 {
	return time.Now().Unix()
}

//
func (ctx *TemplateContext) Base64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

//
func (ctx *TemplateContext) SHA1(value string) string {
	h := sha1.New()
	h.Write([]byte(value))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func executeTemplate(ctx *TemplateContext, tmpl string) (string, error) {
	// variable's syntax could be used inside the template, evaluate vars first
	tmpl = ctx.vars.ApplyTo(tmpl)

	t := template.Must(template.New("value").Parse(tmpl))

	output := bytes.NewBufferString("")

	err := t.Execute(output, ctx)
	if err != nil {
		return "", errors.Wrapf(err, "cannot evaluate value template")
	}
	// t.Execute()

	return output.String(), nil
}
