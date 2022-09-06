package metrics_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/metrics/timestats"
)

func TestAggregate_Get(t *testing.T) {
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
			if got := c.agg.Get(c.fieldID); got != c.exp {
				t.Errorf("exp %v, got %v", c.exp, got)
			}
		})
	}
}
