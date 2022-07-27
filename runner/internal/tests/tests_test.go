package tests_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/tests"
)

func TestRun(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		agg := metricsStub()
		queries := []tests.Input{
			{
				Name: "metrics.Min",
				Metric: func(agg metrics.Aggregate) tests.Value {
					return tests.Value(agg.Min)
				},
				Predicate: tests.GT,
				Value:     tests.Value(50 * time.Millisecond), // expect pass
			},
			{
				Name: "metrics.Max",
				Metric: func(agg metrics.Aggregate) tests.Value {
					return tests.Value(agg.Max)
				},
				Predicate: tests.LT,
				Value:     tests.Value(110 * time.Millisecond), // expect fail
			},
		}
		result := tests.Run(agg, queries)
		fmt.Printf("%+v\n", result)
	})
}

func metricsStub() metrics.Aggregate {
	return metrics.Aggregate{
		Min: 80 * time.Millisecond,
		Max: 120 * time.Millisecond,
		Avg: 100 * time.Millisecond,

		TotalCount:   10,
		FailureCount: 1,
		SuccessCount: 9,
	}
}
