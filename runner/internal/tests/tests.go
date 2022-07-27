package tests

import (
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
)

type Value = time.Duration

type Metric string

const (
	MetricAvg          Metric = "AVG"
	MetricMin          Metric = "MIN"
	MetricMax          Metric = "MAX"
	MetricFailureCount Metric = "FAILURE_COUNT"
	MetricSuccessCount Metric = "SUCCESS_COUNT"
	MetricTotalCount   Metric = "TOTAL_COUNT"
)

type Input struct {
	Name      string
	Metric    func(metrics.Aggregate) Value
	Predicate Predicate
	Value     Value
}

type SuiteResult struct {
	Pass    bool
	Results []SingleResult
}

type SingleResult struct {
	Pass    bool
	Explain string
}

func Run(agg metrics.Aggregate, inputs []Input) SuiteResult {
	allpass := true
	results := make([]SingleResult, len(inputs))
	for i, input := range inputs {
		currentResult := runSingle(agg, input)
		results[i] = currentResult
		if !currentResult.Pass {
			allpass = false
		}
	}
	return SuiteResult{
		Pass:    allpass,
		Results: results,
	}
}

func runSingle(agg metrics.Aggregate, input Input) SingleResult {
	gotMetric := input.Metric(agg)
	comparedValue := input.Value

	return input.Predicate.Apply(gotMetric, comparedValue)
}
