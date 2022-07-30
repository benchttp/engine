package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
)

// Aggregate is an aggregate of metrics resulting from
// from recorded requests.
type Aggregate struct {
	Min, Max, Avg                          time.Duration
	SuccessCount, FailureCount, TotalCount int
	// Median, StdDev            time.Duration
	// Deciles                   map[int]float64
	// StatusCodeDistribution    map[string]int
	// RequestEventsDistribution map[recorder.Event]int
}

// Compute computes and aggregates metrics from the given
// requests records.
func Compute(records []recorder.Record) (agg Aggregate) {
	if len(records) == 0 {
		return
	}

	agg.TotalCount = len(records)
	agg.Min, agg.Max = records[0].Time, records[0].Time
	for _, rec := range records {
		if rec.Error != "" {
			agg.FailureCount++
		}
		if rec.Time < agg.Min {
			agg.Min = rec.Time
		}
		if rec.Time > agg.Max {
			agg.Max = rec.Time
		}
		agg.Avg += rec.Time / time.Duration(len(records))
	}
	agg.SuccessCount = agg.TotalCount - agg.FailureCount
	return
}
