package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

type Funcs struct {
	vars *Vars
}

// NewFuncs creates new Funcs struct that includes initialized Vars.
func NewFuncs(vars *Vars) *Funcs {
	return &Funcs{vars: vars}
}

//
func (ctx *Funcs) CurrentTimestampSec() int64 {
	return time.Now().Unix()
}

//
func (ctx *Funcs) CurrentTimestampISO() string {
	return time.Now().Format(time.RFC3339)
}

// WSSEPasswordDigest returns password digest according to Web Service Security specification.
//
// Password_Digest = Base64 ( SHA-1 ( nonce + created + password ) )
// https://www.oasis-open.org/committees/download.php/13392/wss-v1.1-spec-pr-UsernameTokenProfile-01.htm
func (ctx *Funcs) WSSEPasswordDigest(nonce, created, password string) string {
	h := sha1.New()
	h.Write([]byte(nonce + created + password))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

//
func (ctx *Funcs) Base64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

// SHA1 returns string representation of SHA1 hash bytes
func (ctx *Funcs) SHA1(value string) string {
	// fmt.Println("Calculating SHA1: " + value)
	h := sha1.New()
	io.WriteString(h, value)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type TemplateContext struct {
	funcs  *Funcs
	vars   *Vars
	errors []error
}

// NewTemplateContext creates new processor.
func NewTemplateContext(vars *Vars) *TemplateContext {
	return &TemplateContext{
		funcs: NewFuncs(vars),
		vars:  vars,
	}
}

// HasErrors checks either at least on error happend during processing or not.
func (ctx *TemplateContext) HasErrors() bool {
	return len(ctx.errors) > 0
}

// Error returns all errors throwned during template(s) processing combine in one error
func (ctx *TemplateContext) Error() error {
	msg := ""
	for _, err := range ctx.errors {
		msg = fmt.Sprintln(err.Error())
	}

	return errors.New(msg)
}

// ApplyTo takes value template and evaluates all variables and functions inside it.
func (ctx *TemplateContext) ApplyTo(tmpl string) string {

	// variable's syntax could be used inside the template, evaluate vars first
	tmpl = ctx.vars.ApplyTo(tmpl)

	t, err := template.New("value").Parse(tmpl)
	if err != nil {
		ctx.errors = append(ctx.errors, errors.Wrapf(err, "cannot parse value template"))
		return ""
	}

	output := bytes.NewBufferString("")

	err = t.Execute(output, ctx.funcs)
	if err != nil {
		ctx.errors = append(ctx.errors, errors.Wrapf(err, "cannot evaluate value template"))
		return ""
	}

	return output.String()
}
