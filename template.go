package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
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
func (ctx *TemplateContext) CurrentTimestampSec() int64 {
	return time.Now().Unix()
}

//
func (ctx *TemplateContext) Base64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

//
func (ctx *TemplateContext) SHA1(value string) string {
	// fmt.Println("Calculating SHA1: " + value)
	h := sha1.New()
	io.WriteString(h, value)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type TemplateProcessor struct {
	ctx    *TemplateContext
	vars   *Vars
	errors []error
}

// NewTemplateProcessor creates new processor.
func NewTemplateProcessor(vars *Vars) *TemplateProcessor {
	return &TemplateProcessor{
		ctx:  NewTemplateContext(vars),
		vars: vars,
	}
}

// HasErrors checks either at least on error happend during processing or not.
func (proc *TemplateProcessor) HasErrors() bool {
	return len(proc.errors) > 0
}

// Error returns all errors throwned during template(s) processing combine in one error
func (proc *TemplateProcessor) Error() error {
	msg := ""
	for _, err := range proc.errors {
		msg = fmt.Sprintln(err.Error())
	}

	return errors.New(msg)
}

// ApplyTo takes value template and evaluates all variables and functions inside it.
func (proc *TemplateProcessor) ApplyTo(tmpl string) string {
	// proc.vars.print(os.Stdout)

	// variable's syntax could be used inside the template, evaluate vars first
	tmpl = proc.vars.ApplyTo(tmpl)

	t, err := template.New("value").Parse(tmpl)
	if err != nil {
		proc.errors = append(proc.errors, errors.Wrapf(err, "cannot parse value template"))
		return ""
	}

	output := bytes.NewBufferString("")

	err = t.Execute(output, proc.ctx)
	if err != nil {
		proc.errors = append(proc.errors, errors.Wrapf(err, "cannot evaluate value template"))
		return ""
	}

	return output.String()
}
