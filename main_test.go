package main

import "testing"

func TestConcatURL(t *testing.T) {

	t.Run("open base and closed path", func(t *testing.T) {
		base := "http://example.com"
		path := "/api/v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

	t.Run("closed base and closed path", func(t *testing.T) {
		base := "http://example.com/api/"
		path := "/v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

	t.Run("closed base and open path", func(t *testing.T) {
		base := "http://example.com/api/"
		path := "v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

	t.Run("open base and open path", func(t *testing.T) {
		base := "http://example.com/api"
		path := "v1/example"
		url, _ := concatURL(base, path)
		if url != "http://example.com/api/v1/example" {
			t.Error("Incorrect url. Expected: http://example.com/api/v1/example. Actual: " + url)
		}
	})

}
