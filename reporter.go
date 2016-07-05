package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Reporter interface {
	Report(result TestResult)
	Flush()
}

type ConsoleReporter struct {
	total  int
	failed int
}

func (r *ConsoleReporter) Report(result TestResult) {
	r.total = r.total + 1
	if result.Cause != nil {
		r.failed = r.failed + 1
		r.reportError(result)
	} else {
		r.reportSuccess(result)
	}
}

func (r ConsoleReporter) reportSuccess(result TestResult) {
	c := color.New(color.FgGreen).Add(color.Bold)
	fmt.Printf("[")
	c.Print("PASSED")
	fmt.Printf("] %s\n", result.Case.Description)
}

func (r ConsoleReporter) reportError(result TestResult) {
	c := color.New(color.FgRed).Add(color.Bold)
	fmt.Printf("[")
	c.Print("FAILED")
	fmt.Printf("] %s\n", result.Case.Description)
	lines := strings.Split(result.Cause.Error(), "\n")

	for _, line := range lines {
		fmt.Printf("\t\t%s \n", line)
	}
}

func (r ConsoleReporter) Flush() {
	fmt.Println("~~~ Summary ~~~")
	fmt.Printf("# of test cases : %v\n", r.total)
	fmt.Printf("# Errors: %v\n", r.failed)

	if r.failed > 0 {
		fmt.Println("~~~ Test run FAILURE! ~~~")
		os.Exit(1)
		return
	} // test run failed

	fmt.Println("~~~ Test run SUCCESS ~~~")
	os.Exit(0)
}

// NewConsoleReporter returns new instance of console reporter
func NewConsoleReporter() Reporter {
	return &ConsoleReporter{}
}

type JUnitXMLReporter struct {
	Writer io.Writer
	suits  []*suite
}

type suites struct {
	XMLName string  `xml:"testsuites"`
	Suits   []suite `xml:"testsuite"`
}

type suite struct {
	Name        string `xml:"name,attr"`
	PackageName string `xml:"package,attr"`
	Tests       int    `xml:"tests,attr"`
	Time        uint16 `xml:"time,attr"`
	Cases       []tc   `xml:"testcase"`
}

type tc struct {
	Name    string   `xml:"name,attr"`
	Failure *failure `xml:"failure,omitempty"`
}

type failure struct {
	Message string `xml:"message,attr"`
}

func (r *JUnitXMLReporter) Report(result TestResult) {
	s := r.findSuite(result.Suite.Name)
	if s == nil {
		s = &suite{Name: result.Suite.Name, PackageName: result.Suite.PackageName}
		r.suits = append(r.suits, s)
	}

	testCase := tc{Name: result.Case.Description}
	if result.Cause != nil {
		testCase.Failure = &failure{Message: result.Cause.Error()}
	}
	s.Tests = s.Tests + 1
	s.Cases = append(s.Cases, testCase)
}

func (r *JUnitXMLReporter) findSuite(name string) *suite {
	for _, s := range r.suits {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (r JUnitXMLReporter) Flush() {
	var data []suite
	for _, d := range r.suits {
		data = append(data, *d)
	}
	d, err := xml.Marshal(suites{Suits: data, XMLName: "test22"})
	if err != nil {
		return
	}
	r.Writer.Write(d)
}

func NewJUnitReporter(writer io.Writer) Reporter {
	return &JUnitXMLReporter{Writer: writer}
}
