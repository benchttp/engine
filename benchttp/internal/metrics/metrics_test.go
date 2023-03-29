package metrics_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/benchttp/internal/metrics"
	"github.com/benchttp/engine/benchttp/internal/metrics/timestats"
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

func TestAggregate_MetricOf(t *testing.T) {
	cases := []struct {
		name    string
		fieldID string
		agg     metrics.Aggregate
		exp     metrics.Value
	}{
		{
			name:    "get metrics from nested struct",
			fieldID: "ResponseTimes.Mean",
			agg: metrics.Aggregate{
				ResponseTimes: timestats.TimeStats{
					Mean: 100 * time.Millisecond,
				},
			},
			exp: 100 * time.Millisecond,
		},
		{
			name:    "get metrics from methods",
			fieldID: "RequestSuccessCount",
			agg: metrics.Aggregate{
				Records:         []struct{ ResponseTime time.Duration }{{}, {}, {}},
				RequestFailures: []struct{ Reason string }{{}},
			},
			exp: 2,
		},
		{
			name:    "get metrics from int map",
			fieldID: "StatusCodesDistribution.404",
			agg: metrics.Aggregate{
				StatusCodesDistribution: map[int]int{200: 10, 404: 5},
			},
			exp: 5,
		},
		{
			name:    "get metrics from string map",
			fieldID: "RequestEventTimes.FirstResponseByte.Mean",
			agg: metrics.Aggregate{
				RequestEventTimes: map[string]timestats.TimeStats{
					"FirstResponseByte": {Mean: 100 * time.Millisecond},
				},
			},
			exp: 100 * time.Millisecond,
		},
		{
			name:    "get metrics from slice",
			fieldID: "RequestFailures.1.Reason",
			agg: metrics.Aggregate{
				RequestFailures: []struct{ Reason string }{{"abc"}, {"def"}},
			},
			exp: "def",
		},
		{
			name:    "case insensitive",
			fieldID: "responsetimes.MEAN",
			agg: metrics.Aggregate{
				ResponseTimes: timestats.TimeStats{
					Mean: 100 * time.Millisecond,
				},
			},
			exp: 100 * time.Millisecond,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.agg.MetricOf(metrics.Field(c.fieldID)).Value; got != c.exp {
				t.Errorf("exp %v, got %v", c.exp, got)
			}
		})
	}
}

// helpers

func metricWithValue(v metrics.Value) metrics.Metric {
	return metrics.Metric{Value: v}
}

func assertPanic(t *testing.T) {
	t.Helper()
	if r := recover(); r == nil {
		t.Error("did not panic")
	}
}
