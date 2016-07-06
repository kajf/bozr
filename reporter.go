package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	ID          int    `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	PackageName string `xml:"package,attr"`
	TimeStamp   string `xml:"timestamp,attr"`
	Time        uint16 `xml:"time,attr"`
	HostName    string `xml:"hostname,attr"`

	Tests    int `xml:"tests,attr"`
	Failures int `xml:"failures,attr"`
	Errors   int `xml:"errors,attr"`

	Properties properties `xml:"properties"`
	Cases      []tc       `xml:"testcase"`

	SystemOut string `xml:"system-out"`
	SystemErr string `xml:"system-err"`
}

type properties struct {
}

type tc struct {
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Time      uint16   `xml:"time,attr"`
	Failure   *failure `xml:"failure,omitempty"`
}

type failure struct {
	Type string `xml:"type,attr"`
}

func (r *JUnitXMLReporter) Report(result TestResult) {
	s := r.findSuite(result.Suite.Name)
	if s == nil {
		s = &suite{
			ID:          0,
			Name:        result.Suite.Name,
			PackageName: result.Suite.PackageName,
			TimeStamp:   time.Now().UTC().Format("2006-01-02T15:04:05"),
			HostName:    "test",
		}
		r.suits = append(r.suits, s)
	}

	testCase := tc{Name: result.Case.Description}
	if result.Cause != nil {
		testCase.Failure = &failure{Type: result.Cause.Error()}
		s.Failures = s.Failures + 1
	}
	s.Tests = s.Tests + 1
	s.ID = s.ID + 1
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
