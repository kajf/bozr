package main

import (
	"errors"
	"fmt"
	"mime"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

// ResponseExpectation is an interface to any validation
// that needs to be performed on response.
type ResponseExpectation interface {
	check(resp *Response) error
}

// StatusCodeExpectation validates response HTTP code.
type StatusCodeExpectation struct {
	statusCode int
}

func (e StatusCodeExpectation) check(resp *Response) error {
	if resp.http.StatusCode != e.statusCode {
		return fmt.Errorf("Unexpected Status Code. Expected: %d, Actual: %d", e.statusCode, resp.http.StatusCode)
	}
	return nil
}

// BodySchemaExpectation validates response body against schema.
// Content-Type header is used to identify either json schema or xsd is applied.
type BodySchemaExpectation struct {
	schemaURI string
}

func (e BodySchemaExpectation) check(resp *Response) error {
	contentType, _, _ := mime.ParseMediaType(resp.http.Header.Get("content-type"))

	if contentType == "application/json" {
		return e.checkJSON(resp)
	}

	return fmt.Errorf("Unsupported content type: %s", contentType)
}

var jsonSchemaCache sync.Map

func (e BodySchemaExpectation) checkJSON(resp *Response) error {
	var safeSchema, ok = jsonSchemaCache.Load(e.schemaURI)

	if !ok {
		debug.Printf("Loading schema %s", e.schemaURI)

		schemaLoader := gojsonschema.NewReferenceLoader(e.schemaURI)
		schema, err := schemaLoader.LoadJSON()

		if err != nil {
			return fmt.Errorf("failed to load schema: %s", err)
		}

		jsonSchemaCache.Store(e.schemaURI, schema)
		safeSchema = schema
	}

	schemaLoader := gojsonschema.NewGoLoader(safeSchema)
	documentLoader := gojsonschema.NewStringLoader(string(resp.body))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("failed to load schema file: %s", err)
	}

	if !result.Valid() {
		msg := "Unexpected Body Schema:\n"
		for _, desc := range result.Errors() {
			msg = fmt.Sprintf(msg+"%s\n", desc)
		}
		return errors.New(msg)
	}

	return nil
}

// BodyExpectation validates values under a certain path in a body.
// Applies to json and xml.
type BodyExpectation struct {
	pathExpectations map[string]interface{}
}

func (e BodyExpectation) check(resp *Response) error {

	for pathStr, expectedValue := range e.pathExpectations {

		err := responseBodyPathCheck(resp, bodyExpectationItem{Path: pathStr, ExpectedValue: expectedValue}, checkExpectedPath)
		if err != nil {
			return err
		}
	}

	return nil
}

type bodyExpectationItem struct {
	Path          string
	ExpectedValue interface{}
}

func checkExpectedPath(m interface{}, pathItem interface{}) string {

	if expectationItem, ok := pathItem.(bodyExpectationItem); ok {

		ok, err := SearchByPath(m, expectationItem.ExpectedValue, expectationItem.Path)
		if !ok {
			return fmt.Sprintf("Expected value %#v on path %#v is not found", expectationItem.ExpectedValue, expectationItem.Path)
		}
		if err != nil {
			return err.Error()
		}

		return ""
	}

	return fmt.Sprintf("Path Item: %v is invalid for expectation check", pathItem)
}

// HeaderExpectation validates one header in a response.
type HeaderExpectation struct {
	Name        string
	Value       string
	ValueParser func(string) string
}

func (e HeaderExpectation) check(resp *Response) error {
	value := resp.http.Header.Get(e.Name)
	if e.ValueParser != nil {
		value = e.ValueParser(value)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("Missing header. Expected \"%s: %s\"", e.Name, e.Value)
	}
	if e.Value != "" && e.Value != value {
		return fmt.Errorf("Unexpected header. Expected \"%s: %s\". Actual \"%s: %s\"", e.Name, e.Value, e.Name, value)
	}
	return nil
}

// ContentTypeExpectation validates media type returned in the Content-Type header.
// Encoding information is excluded from matching value.
// E.g. "application/json;charset=utf-8" header transformed to "application/json" media type.
type ContentTypeExpectation struct {
	Value string
}

func (e ContentTypeExpectation) check(resp *Response) error {
	parser := func(value string) string {
		contentType, _, _ := mime.ParseMediaType(value)
		return contentType
	}

	headerCheck := HeaderExpectation{"content-type", e.Value, parser}
	return headerCheck.check(resp)
}

// AbsentExpectation validates paths are absent in response body
type AbsentExpectation struct {
	paths []string
}

func (e AbsentExpectation) check(resp *Response) error {

	for _, pathStr := range e.paths {
		err := responseBodyPathCheck(resp, pathStr, checkAbsentPath)
		if err != nil {
			return err
		}
	}

	return nil
}

type pathCheckFunc func(m interface{}, pathItem interface{}) string

func responseBodyPathCheck(resp *Response, pathItem interface{}, checkPath pathCheckFunc) error {

	m, err := resp.Body() // cached
	if err != nil {
		str := "Can't parse response body to Map." // TODO specific message for functions
		str += " " + err.Error()

		return errors.New(str)
	}

	str := checkPath(m, pathItem)
	if str != "" {
		return errors.New(str)
	}

	return nil
}

func checkAbsentPath(m interface{}, pathItem interface{}) string {

	if pathStr, ok := pathItem.(string); ok {

		searchResult := Search(m, pathStr)
		if len(searchResult) > 0 {
			return fmt.Sprintf("Value expected to be absent was found: %v, path: %v", searchResult, pathStr)
		}

		return ""
	}

	return fmt.Sprintf("Path Item: %v is invalid for absence check", pathItem)
}
