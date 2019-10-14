package main

import (
	"errors"
	"github.com/fatih/color"
	"strings"
	"sync"
	"testing"
)

func TestConsoleReporterReport_ErrorAfterPassedExp_Reported(t *testing.T) {
	// given
	reportedError := "test err"
	results := []TestResult{
		{
			Traces: []*CallTrace{
				{
					ExpDesc:    map[string]bool{"Status code is 200": false},
					ErrorCause: errors.New(reportedError),
				},
			},
		}, // error without expectation description
	}

	writer := MockWriter{
		expectedWriting: reportedError,
	}

	reporter := &ConsoleReporter{ExitCode: 0, Writer: &writer, ioMutex: &sync.Mutex{}, LogHTTP: false}

	color.Output = &writer // prevent stdout and invalid test result parsing in IDE (reacts on words 'FAILED')

	// when
	reporter.Report(results)

	// then
	if !writer.passed() {
		t.Errorf("Expected writing %s was not met in %s", writer.expectedWriting, writer.actualWriting)
	}
}

type MockWriter struct {
	expectedWriting string
	actualWriting   string
}

func (mw *MockWriter) Write(p []byte) (n int, err error) {

	in := string(p)
	mw.actualWriting += in

	//os.Stdout.Write(p)

	return len(p), nil
}

func (mw *MockWriter) passed() bool {
	return strings.Contains(mw.actualWriting, mw.expectedWriting)
}
