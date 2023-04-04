package testsuite

import (
	"fmt"

	"github.com/benchttp/engine/benchttp/metrics"
)

type Case struct {
	Name      string
	Field     metrics.Field
	Predicate Predicate
	Target    metrics.Value
}


type CaseResult struct {
	Input   Case
	Pass    bool
	Got     metrics.Value
	Summary string
}

type Result struct {
	Pass    bool
	OfCases []CaseResult
}

func Run(agg metrics.Aggregate, cases []Case) Result {
	allpass := true
	results := make([]CaseResult, len(cases))
	for i, input := range cases {
		currentResult := runTestCase(agg, input)
		results[i] = currentResult
		if !currentResult.Pass {
			allpass = false
		}
	}
	return Result{
		Pass:    allpass,
		OfCases: results,
	}
}

func runTestCase(agg metrics.Aggregate, c Case) CaseResult {
	gotMetric := agg.MetricOf(c.Field)
	tarMetric := metrics.Metric{Field: c.Field, Value: c.Target}
	comparisonResult := gotMetric.Compare(tarMetric)

	return CaseResult{
		Input: c,
		Pass:  c.Predicate.match(comparisonResult),
		Got:   gotMetric.Value,
		Summary: fmt.Sprintf(
			"want %s %s %v, got %v",
			c.Field, c.Predicate.symbol(), c.Target, gotMetric.Value,
		),
	}
}
