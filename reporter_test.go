package main

import (
	"testing"
)

func TestJUnitReporterEmptyResults(t *testing.T) {
	// given

	// when
	NewJUnitReporter("").Report([]TestResult{})

	// then
	// no nil pointer panic
}
