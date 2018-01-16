package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

// Reporter is used to write down test results using particular formats and outputs
type Reporter interface {
	Init()

	Report(result []TestResult)

	Flush()
}

// ConsoleReporter is a simple reporter that outputs everything to the StdOut.
type ConsoleReporter struct {
	ExitCode   int
	LogHTTP    bool
	Writer     io.Writer
	IndentSize int

	execFrame *TimeFrame

	// to prevent collisions while working with StdOut
	ioMutex *sync.Mutex

	total   int
	failed  int
	skipped int
}

func (r *ConsoleReporter) Init() {
	r.execFrame = &TimeFrame{Start: time.Now()}
}

const (
	defaultIndentSize = 4
	caretIcon         = "\u2514" // ↳
)

type status struct {
	Icon  string
	Label string
	Color color.Attribute
}

const (
	outputLabel int = iota
	outputIcon
)

var (
	statusPassed  = status{Icon: "\u221A", Label: "PASSED", Color: color.FgGreen} // ✔
	statusFailed  = status{Icon: "\u00D7", Label: "FAILED", Color: color.FgRed}   // ✘
	statusSkipped = status{Icon: "", Label: "SKIPPED", Color: color.FgYellow}
)

func (r *ConsoleReporter) verbose() bool {
	return r.LogHTTP
}

func (r *ConsoleReporter) StartLine() {
	r.Writer.Write([]byte("\n"))
	r.Writer.Write([]byte(strings.Repeat(" ", r.IndentSize)))
}

func (r *ConsoleReporter) Indent() {
	r.IndentSize = r.IndentSize + defaultIndentSize
}

func (r *ConsoleReporter) Unindent() {
	r.IndentSize = r.IndentSize - defaultIndentSize
}

func (r *ConsoleReporter) Report(results []TestResult) {
	r.ioMutex.Lock()

	if len(results) == 0 {
		r.ioMutex.Unlock()
		return
	}

	// suite
	suite := results[0].Suite

	r.StartLine()
	r.Write(suite.FullName())

	for _, result := range results {

		r.total = r.total + 1

		r.Indent()

		r.StartLine()
		r.Write(caretIcon).Write(" ")

		if result.Skipped {
			r.WriteStatus(statusSkipped, outputLabel).Write(" ").Write(result.Case.Name)

			skippedFg := color.New(color.FgHiYellow)
			skippedFg.Print(" (")
			skippedFg.Print(result.SkippedMsg)
			skippedFg.Print(") ")

			r.skipped = r.skipped + 1
			r.Unindent()

			continue
		}

		if result.hasError() {
			r.WriteStatus(statusFailed, outputLabel)
			r.failed = r.failed + 1
		} else {
			r.WriteStatus(statusPassed, outputLabel)
		}

		r.Write(" ").Write(result.Case.Name)
		r.Write(" [").Write(result.ExecFrame.Duration().Round(time.Millisecond)).Write("]")

		if result.hasError() || r.LogHTTP {
			for _, trace := range result.Traces {
				r.Indent()

				r.StartLine()
				r.Write(trace.RequestMethod).Write(" ").Write(trace.RequestURL).Write(" [").Write(trace.ExecFrame.Duration().Round(time.Millisecond)).Write("]")

				for exp, failed := range trace.ExpDesc {
					r.Indent()
					r.StartLine()

					if failed {
						r.WriteStatus(statusFailed, outputIcon)
					} else {
						r.WriteStatus(statusPassed, outputIcon)
					}

					r.Write(" ").WriteMultiline(exp, r.Write)

					r.Unindent()
				}

				if r.LogHTTP {
					r.StartLine()

					r.Indent()
					{
						dump := trace.RequestDump
						if len(dump) > 0 {
							r.WriteMultiline(dump, r.WriteDimmed)
							r.StartLine()
						}

						dump = trace.ResponseDump
						if len(dump) > 0 {
							r.WriteMultiline(trace.ResponseDump, r.WriteDimmed)
							r.StartLine()
						}
					}
					r.Unindent()
				}

				r.Unindent()
			}
		}

		r.Unindent()
	}

	r.StartLine()

	r.ioMutex.Unlock()
}

func (r ConsoleReporter) WriteDimmed(content interface{}) ConsoleReporter {
	c := color.New(color.FgHiBlack)
	c.Print(content)
	return r
}

func (r ConsoleReporter) WriteMultiline(content string, writer func(content interface{}) ConsoleReporter) ConsoleReporter {
	for i, line := range strings.Split(content, "\n") {
		if i > 0 {
			r.StartLine()
		}
		writer(line)
	}
	return r
}

func (r ConsoleReporter) Write(content interface{}) ConsoleReporter {
	r.Writer.Write([]byte(fmt.Sprintf("%v", content)))
	return r
}

func (r ConsoleReporter) WriteStatus(status status, output int) ConsoleReporter {
	c := color.New(status.Color).Add(color.Bold)
	var val string

	if output == outputIcon {
		val = status.Icon
	}

	if output == outputLabel {
		val = status.Label
	}

	c.Print(val)
	return r
}

func (r ConsoleReporter) Flush() {
	r.ioMutex.Lock()
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

	fmt.Fprintf(w, "Start time:\t %s\n", start.Round(time.Millisecond))
	fmt.Fprintf(w, "End time:\t %s\n", end.Round(time.Millisecond))
	fmt.Fprintf(w, "Duration:\t %s\n", end.Sub(start).Round(time.Millisecond))

	w.Flush()
	fmt.Println()
	r.ioMutex.Unlock()
}

// NewConsoleReporter returns new instance of console reporter
func NewConsoleReporter(logHTTP bool) Reporter {
	return &ConsoleReporter{ExitCode: 0, ioMutex: &sync.Mutex{}, Writer: os.Stdout, LogHTTP: logHTTP}
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

		if result.hasError() {
			errType := "FailedExpectation"
			errMsg := result.Error()

			errIndex := 0
			errRespDump := ""
			for index, trace := range result.Traces {
				if trace.hasError() {
					errIndex = index
					errRespDump = string(trace.ResponseDump)
				}
			}

			errDetails := fmt.Sprintf("On Call %d - %s\n\n%s", errIndex, errMsg, errRespDump)

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
