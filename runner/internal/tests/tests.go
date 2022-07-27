package tests

import (
	"github.com/benchttp/engine/runner/internal/metrics"
)

type Value = int

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
