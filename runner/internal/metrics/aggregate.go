package metrics

import (
	"errors"

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
func Compute(records []recorder.Record) (agg Aggregate, err error) {
	if len(records) == 0 {
		return
	}

	agg.TotalCount = len(records)
	agg.SuccessCount = agg.TotalCount - agg.FailureCount

	errs := []error{}
	agg.ResponseTimes, errs = timestats.Compute(records)

	if len(errs) > 0 {
		return agg, errors.New("could not compute time stats")
	}

	return agg, nil
}
