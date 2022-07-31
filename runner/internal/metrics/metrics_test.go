package metrics_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
)

func TestMetric_Compare(t *testing.T) {
	const (
		base = 100
		more = base + 1
		less = base - 1
	)

	testcases := []struct {
		label        string
		baseMetric   metrics.Metric
		targetMetric metrics.Metric
		expResult    metrics.ComparisonResult
		expPanic     bool
	}{
		{
			label:        "base equals target",
			baseMetric:   metricWithValue(base),
			targetMetric: metricWithValue(base),
			expResult:    metrics.EQ,
			expPanic:     false,
		},
		{
			label:        "base superior to target",
			baseMetric:   metricWithValue(base),
			targetMetric: metricWithValue(less),
			expResult:    metrics.SUP,
			expPanic:     false,
		},
		{
			label:        "base inferior to target",
			baseMetric:   metricWithValue(base),
			targetMetric: metricWithValue(more),
			expResult:    metrics.INF,
			expPanic:     false,
		},
		{
			label:        "panics with different type",
			baseMetric:   metricWithValue(base),
			targetMetric: metricWithValue(base * time.Millisecond),
			expResult:    0, // irrelevant, should panic
			expPanic:     true,
		},
		{
			label:        "panics with different type",
			baseMetric:   metricWithValue(1.23),
			targetMetric: metricWithValue(1.23),
			expResult:    0, // irrelevant, should panic
			expPanic:     true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			if tc.expPanic {
				defer assertPanic(t)
			}

			result := tc.baseMetric.Compare(tc.targetMetric)

			if !tc.expPanic && result != tc.expResult {
				t.Errorf(
					"\nexp %v.Compare(%v) == %v, got %v",
					tc.baseMetric, tc.targetMetric, tc.expResult, result)
			}
		})
	}
}

func metricWithValue(v metrics.Value) metrics.Metric {
	return metrics.Metric{Value: v}
}

func assertPanic(t *testing.T) {
	t.Helper()
	if r := recover(); r == nil {
		t.Error("did not panic")
	}
}
