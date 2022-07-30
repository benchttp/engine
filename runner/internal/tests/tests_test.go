package tests_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/tests"
)

func TestRun(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		agg := metricsStub()
		queries := []tests.Case{
			{
				Name:      "minimum response time is 50ms", // succeeding case
				Source:    metrics.ResponseTimeMax,
				Predicate: tests.GT,
				Target:    metrics.Value(50 * time.Millisecond),
			},
			{
				Name:      "maximum response time is 110ms", // failing case
				Source:    metrics.ResponseTimeMax,
				Predicate: tests.LT,
				Target:    metrics.Value(110 * time.Millisecond),
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
