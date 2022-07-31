package tests

import (
	"fmt"

	"github.com/benchttp/engine/runner/internal/metrics"
)

type Case struct {
	Name      string
	Source    metrics.Source
	Predicate Predicate
	Target    metrics.Value
}

type SuiteResult struct {
	Pass    bool
	Results []CaseResult
}

type CaseResult struct {
	Input   Case
	Pass    bool
	Summary string
}

func Run(agg metrics.Aggregate, cases []Case) SuiteResult {
	allpass := true
	results := make([]CaseResult, len(cases))
	for i, input := range cases {
		currentResult := runTestCase(agg, input)
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

func runTestCase(agg metrics.Aggregate, c Case) CaseResult {
	gotMetric := agg.MetricOf(c.Source)
	tarMetric := metrics.Metric{Source: c.Source, Value: c.Target}
	comparisonResult := gotMetric.Compare(tarMetric)

	return CaseResult{
		Input: c,
		Pass:  c.Predicate.match(comparisonResult),
		Summary: fmt.Sprintf(
			"want %s %s %v, got %v",
			c.Source, c.Predicate.symbol(), c.Target, gotMetric.Value,
		),
	}
}
