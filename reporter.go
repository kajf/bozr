package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
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

// JUnitXMLReporter produces separate xml file for each test sute
type JUnitXMLReporter struct {
	// output directory
	OutPath string

	// current suite
	// when suite is being changed, flush previous one
	suite *suite
}

type suite struct {
	XMLName     string `xml:"testsuite"`
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
	// not clear what type is but it's required
	Type    string `xml:"type,attr"`
	Message string `xml:"message,attr"`
}

func (r *JUnitXMLReporter) Report(result TestResult) {
	if r.suite == nil {
		r.suite = newSuite(result)
	}

	if r.suite.Name != result.Suite.Name {
		r.flushSuite()
		r.suite = newSuite(result)
	}

	testCase := tc{Name: result.Case.Description, ClassName: result.Case.Description}
	if result.Cause != nil {
		testCase.Failure = &failure{Type: result.Cause.Error(), Message: result.Cause.Error()}
		r.suite.Failures = r.suite.Failures + 1
	}
	r.suite.Tests = r.suite.Tests + 1
	r.suite.ID = r.suite.ID + 1
	r.suite.Cases = append(r.suite.Cases, testCase)
}

func (r JUnitXMLReporter) flushSuite() {
	fileName := strings.Replace(filepath.ToSlash(r.suite.PackageName), "/", "_", -1) + r.suite.Name + ".xml"
	fp := filepath.Join(r.OutPath, fileName)
	err := os.MkdirAll(r.OutPath, 0777)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(fp)
	if err != nil {
		panic(err)
	}

	data, err := xml.Marshal(r.suite)
	if err != nil {
		panic(err)
	}

	f.Write(data)
}

func newSuite(result TestResult) *suite {
	return &suite{
		ID:          0,
		Name:        result.Suite.Name,
		PackageName: result.Suite.PackageName,
		TimeStamp:   time.Now().UTC().Format("2006-01-02T15:04:05"),
		HostName:    "test",
	}
}

func (r JUnitXMLReporter) Flush() {
	r.flushSuite()
}

func NewJUnitReporter(outdir string) Reporter {
	return &JUnitXMLReporter{OutPath: outdir}
}
