package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

// Aggregate is an aggregate of metrics resulting from
// from recorded requests.
type Aggregate struct {
	TimeStats                              timestats.TimeStats
	SuccessCount, FailureCount, TotalCount int
	// Median, StdDev            time.Duration
	// Deciles                   map[int]float64
	// StatusCodeDistribution    map[string]int
	// RequestEventsDistribution map[recorder.Event]int
}

// Compute computes and aggregates metrics from the given
// requests records.
func Compute(records []recorder.Record) (agg Aggregate, err error) {
	if len(records) == 0 {
		return
	}

	times := make([]time.Duration, len(records))
	for i, v := range records {
		times[i] = v.Time
	}

	var errs []string
	agg.TimeStats, errs = timestats.Compute(times)

	agg.TotalCount = len(records)
	for _, rec := range records {
		if rec.Error != "" {
			agg.FailureCount++
		}
	}
	agg.SuccessCount = agg.TotalCount - agg.FailureCount

	if len(outErrs) > 0 {
		return agg, outErrs
	}

	return agg, nil
}
