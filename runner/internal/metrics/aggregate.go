package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

// Aggregate is an aggregate of metrics resulting from
// from recorded requests.
type Aggregate struct {
	ResponseTimes          timestats.TimeStats
	StatusCodeDistribution map[string]int
	RequestEventTimes      map[string]timestats.TimeStats
	RequestFailures        []struct {
		Reason string
	}
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

	times := make([]time.Duration, len(records))
	for i, v := range records {
		times[i] = v.Time
	}

	agg.ResponseTimes = timestats.Compute(times)

	for _, rec := range records {
		if rec.Error != "" {
			agg.RequestFailures = append(agg.RequestFailures, struct{ Reason string }{rec.Error})
		}
	}

	var statusCodeDistributionErrs []error
	agg.StatusCodeDistribution, statusCodeDistributionErrs = ComputeStatusCodesDistribution(records)
	errs = append(errs, statusCodeDistributionErrs...)

	agg.RequestEventTimes = ComputeRequestEventTimes(records)

	if len(errs) > 0 {
		return agg, errs
	}

	return agg, nil
}
