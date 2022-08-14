package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

// Aggregate is an aggregate of metrics resulting from
// from recorded requests.
type Aggregate struct {
	SuccessCount, FailureCount, TotalCount int
	ResponseTimes          timestats.TimeStats
	StatusCodeDistribution map[string]int
	RequestEventTimes      map[string]timestats.TimeStats
}

// MetricOf returns the Metric for the given field in Aggregate.
//
// It panics if field is not a known field.
func (agg Aggregate) MetricOf(field Field) Metric {
	return Metric{Field: field, Value: field.valueIn(agg)}
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
