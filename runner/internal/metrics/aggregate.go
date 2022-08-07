package metrics

import (
	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

// Aggregate is an aggregate of metrics resulting from
// from recorded requests.
type Aggregate struct {
	ResponseTimes                          timestats.TimeStats
	SuccessCount, FailureCount, TotalCount int
	// RequestEventsDistribution map[recorder.Event]int
}

// MetricOf returns the Metric for the given field in Aggregate.
//
// It panics if field is not a known field.
func (agg Aggregate) MetricOf(field Field) Metric {
	return Metric{Field: field, Value: field.valueIn(agg)}
}

// Compute computes and aggregates metrics from the given
// requests records.
func Compute(records []recorder.Record) (agg Aggregate, errs []error) {
	if len(records) == 0 {
		return
	}

	agg.TotalCount = len(records)
	agg.SuccessCount = agg.TotalCount - agg.FailureCount

	var reponseTimesErrs []error
	agg.ResponseTimes, reponseTimesErrs = timestats.Compute(records)

	errs = append(errs, reponseTimesErrs...)

	if len(errs) > 0 {
		return agg, errs
	}

	return agg, nil
}
