package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

type Reporter interface {
	Init()

	Report(result []TestResult)

	Flush()
}

type ConsoleReporter struct {
	ExitCode int

	execFrame *TimeFrame

	total   int
	failed  int
	skipped int
}

func (r *ConsoleReporter) Init() {
	r.execFrame = &TimeFrame{Start: time.Now()}
}

func (r *ConsoleReporter) Report(results []TestResult) {

	for _, result := range results {
		r.total = r.total + 1

		if result.Skipped {
			r.reportSkipped(result)
			r.skipped = r.skipped + 1
			continue
		}

		if result.Error != nil {
			r.failed = r.failed + 1
			r.reportError(result)
		} else {
			r.reportSuccess(result)
		}
	}
}

func (r ConsoleReporter) reportSuccess(result TestResult) {
	c := color.New(color.FgGreen).Add(color.Bold)
	fmt.Printf("[")
	c.Print("PASSED")
	fmt.Printf("]  %s - %s \t%s\n", result.Suite.FullName(), result.Case.Name, result.ExecFrame.Duration())
}

func (r ConsoleReporter) reportSkipped(result TestResult) {
	c := color.New(color.FgYellow).Add(color.Bold)
	fmt.Printf("[")
	c.Print("SKIPPED")
	fmt.Printf("] %s - %s", result.Suite.FullName(), result.Case.Name)
	if result.SkippedMsg != "" {
		reasonColor := color.New(color.FgMagenta)
		reasonColor.Printf("\t (%s)", result.SkippedMsg)
	}

	fmt.Printf("\n")
}

func (r ConsoleReporter) reportError(result TestResult) {
	c := color.New(color.FgRed).Add(color.Bold)
	fmt.Printf("[")
	c.Print("FAILED")
	fmt.Printf("]  %s - %s \n", result.Suite.FullName(), result.Case.Name)
	lines := strings.Split(result.Error.Cause.Error(), "\n")

	for _, line := range lines {
		fmt.Printf("\t\t%s \n", line)
	}
}

func (r ConsoleReporter) Flush() {
	r.execFrame.End = time.Now()

	overall := "PASSED"
	if r.failed != 0 {
		overall = "FAILED"
	}

	fmt.Println()
	fmt.Println("Test Run Summary")
	fmt.Println("-------------------------------")

	w := tabwriter.NewWriter(os.Stdout, 4, 2, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "Overall result:\t %s\n", overall)

	fmt.Fprintf(w, "Test count:\t %d\n", r.total)

	fmt.Fprintf(w, "Passed:\t %d \n", r.total-r.failed-r.skipped)
	fmt.Fprintf(w, "Failed:\t %d \n", r.failed)
	fmt.Fprintf(w, "Skipped:\t %d \n", r.skipped)

	start := r.execFrame.Start
	end := r.execFrame.End

	fmt.Fprintf(w, "Start time:\t %s\n", start)
	fmt.Fprintf(w, "End time:\t %s\n", end)
	fmt.Fprintf(w, "Duration:\t %s\n", end.Sub(start).String())

	w.Flush()
	fmt.Println()
}

// NewConsoleReporter returns new instance of console reporter
func NewConsoleReporter() Reporter {
	return &ConsoleReporter{ExitCode: 0}
}

// JUnitXMLReporter produces separate xml file for each test sute
type JUnitXMLReporter struct {
	// output directory
	OutPath string
}

func (r *JUnitXMLReporter) Init() {
	// nothing to do here
}

type suite struct {
	XMLName     string  `xml:"testsuite"`
	ID          int     `xml:"id,attr"`
	Name        string  `xml:"name,attr"`
	PackageName string  `xml:"package,attr"`
	TimeStamp   string  `xml:"timestamp,attr"`
	Time        float64 `xml:"time,attr"`
	HostName    string  `xml:"hostname,attr"`

	Tests    int `xml:"tests,attr"`
	Failures int `xml:"failures,attr"`
	Errors   int `xml:"errors,attr"`
	Skipped  int `xml:"skipped,attr"`

	Properties properties `xml:"properties"`
	Cases      []tc       `xml:"testcase"`

	SystemOut string `xml:"system-out"`
	SystemErr string `xml:"system-err"`

	fullName string
}

type properties struct {
}

type tc struct {
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Time      float64  `xml:"time,attr"`
	Failure   *failure `xml:"failure,omitempty"`
	Skipped   *skipped `xml:"skipped,omitempty"`
}

type failure struct {
	// not clear what type is but it's required
	Type    string `xml:"type,attr"`
	Message string `xml:"message,attr"`
	Details string `xml:",chardata"`
}

type skipped struct {
	Message string `xml:"message,attr"`
}

func (r *JUnitXMLReporter) Report(results []TestResult) {

	var suiteResult *suite
	var suiteTimeFrame TimeFrame
	for _, result := range results {

		if suiteResult == nil {
			suiteResult = &suite{
				ID:          0,
				Name:        result.Suite.Name,
				PackageName: result.Suite.PackageName(),
				TimeStamp:   time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				fullName:    result.Suite.FullName(),
				HostName:    "localhost",
			}

			suiteTimeFrame = result.ExecFrame
		}

		testCase := tc{
			Name:      result.Case.Name,
			ClassName: suiteResult.fullName,
			Time:      result.ExecFrame.Duration().Seconds(),
		}

		if result.Error != nil {
			errType := "FailedExpectation"
			errMsg := result.Error.Cause.Error()
			errDetails := fmt.Sprintf("%s\n\n%s", errMsg, result.Error.Resp.ToString())

			testCase.Failure = &failure{
				Type:    errType,
				Message: errMsg,
				Details: errDetails,
			}

			suiteResult.Failures = suiteResult.Failures + 1
		}

		if result.Skipped {
			suiteResult.Skipped = suiteResult.Skipped + 1
			testCase.Skipped = &skipped{Message: result.SkippedMsg}
		}

		suiteResult.Tests = suiteResult.Tests + 1
		suiteResult.ID = suiteResult.ID + 1
		suiteResult.Cases = append(suiteResult.Cases, testCase)

		suiteTimeFrame.Extend(result.ExecFrame)
		suiteResult.Time = suiteTimeFrame.Duration().Seconds()
	}

	r.flushSuite(suiteResult)
}

func (r JUnitXMLReporter) flushSuite(suite *suite) {
	if suite == nil {
		return
	}

	fileName := suite.fullName + ".xml"
	fp := filepath.Join(r.OutPath, fileName)
	err := os.MkdirAll(r.OutPath, 0777)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(fp)
	if err != nil {
		panic(err)
	}

	data, err := xml.Marshal(suite)
	if err != nil {
		panic(err)
	}

	f.Write(data)
}

func (r JUnitXMLReporter) Flush() {

}

func NewJUnitReporter(outdir string) Reporter {
	return &JUnitXMLReporter{OutPath: outdir}
}

// MultiReporter broadcasts events to another reporters.
type MultiReporter struct {
	Reporters []Reporter
}

func (r MultiReporter) Report(results []TestResult) {
	for _, reporter := range r.Reporters {
		reporter.Report(results)
	}
}

func (r MultiReporter) Init() {
	for _, reporter := range r.Reporters {
		reporter.Init()
	}
}

func (r MultiReporter) Flush() {
	for _, reporter := range r.Reporters {
		reporter.Flush()
	}
}

// NewMultiReporter creates new reporter that broadcasts events to another reporters.
func NewMultiReporter(reporters ...Reporter) Reporter {
	return &MultiReporter{Reporters: reporters}
}
