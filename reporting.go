package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Reporter interface {
	Report(result TestResult)
}

type ConsoleReporter struct {
}

func (r ConsoleReporter) Report(result TestResult) {
	fmt.Println(result.Cause)
	if result.Cause != nil {
		r.reportError(result)
	} else {
		r.reportSuccess(result)
	}
}

func (r ConsoleReporter) reportSuccess(result TestResult) {
	c := color.New(color.FgGreen).Add(color.Bold)
	c.Printf("PASSED: %s\n", result.Case.Description)
}

func (r ConsoleReporter) reportError(result TestResult) {
	c := color.New(color.FgRed).Add(color.Bold)
	c.Printf("FAILED: %s\n", result.Case.Description)
	lines := strings.Split(result.Cause.Error(), "\n")

	for _, line := range lines {
		fmt.Printf("\t%s \n", line)
	}
}

// NewConsoleReporter returns new instance of console reporter
func NewConsoleReporter() Reporter {
	return &ConsoleReporter{}
}
