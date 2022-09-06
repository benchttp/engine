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
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.agg.Get(c.fieldID); got != c.exp {
				t.Errorf("exp %v, got %v", c.exp, got)
			}
		})
	}
}
