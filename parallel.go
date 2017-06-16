package main

import (
	"sync"
)

// RunSuiteFunc describes particular test suite execution. Passed here to deleniate parallelism from suite execution logic
type RunSuiteFunc func(suite TestSuite) []TestResult

// RunParallel starts parallel routines to execute test suites received from loader channel
func RunParallel(loader <-chan TestSuite, reporter Reporter, runSuite RunSuiteFunc, numRoutines int) {

	resultConsumer := make(chan []TestResult)

	var wg sync.WaitGroup
	wg.Add(numRoutines)

	for i := 0; i < numRoutines; i++ {
		go runSuites(loader, resultConsumer, &wg, runSuite)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(resultConsumer)
	}()

	for {
		results, more := <-resultConsumer

		reporter.Report(results)

		if !more {
			break
		}
	}

	reporter.Flush()
}

func runSuites(loader <-chan TestSuite, resultConsumer chan<- []TestResult, wg *sync.WaitGroup, runSuite RunSuiteFunc) {

	for suite := range loader {
		resultConsumer <- runSuite(suite)
	}

	wg.Done()
}
