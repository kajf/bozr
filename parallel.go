package main

import (
	"sync"
)

// RunSuiteFunc describes particular test suite execution. Passed here to deleniate parallelism from suite execution logic
type RunSuiteFunc func(requestConfig *RequestConfig, rewriteConfig *RewriteConfig, suite TestSuite) []TestResult

// RunParallel starts parallel routines to execute test suites received from loader channel
func RunParallel(runConfig *RunConfig) {

	resultConsumer := make(chan []TestResult)

	var wg sync.WaitGroup
	wg.Add(runConfig.numRoutines)

	for i := 0; i < runConfig.numRoutines; i++ {
		go runSuites(&SuiteConfig{
			requestConfig:  runConfig.requestConfig,
			rewriteConfig:  runConfig.rewriteConfig,
			loader:         runConfig.loader,
			resultConsumer: resultConsumer,
			waitGroup:      &wg,
			runner:         runConfig.runSuite,
		})
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(resultConsumer)
	}()

	for {
		results, more := <-resultConsumer

		runConfig.reporter.Report(results)

		if !more {
			break
		}
	}

	runConfig.reporter.Flush()
}

type SuiteConfig struct {
	requestConfig  *RequestConfig
	rewriteConfig  *RewriteConfig
	loader         <-chan TestSuite
	resultConsumer chan []TestResult
	waitGroup      *sync.WaitGroup
	runner         RunSuiteFunc
}

func runSuites(cfg *SuiteConfig) {

	for suite := range cfg.loader {
		cfg.resultConsumer <- runSuite(cfg.requestConfig, cfg.rewriteConfig, suite)
	}

	cfg.waitGroup.Done()
}
