package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

// Aggregate is an aggregate of metrics resulting from
// from recorded requests.
type Aggregate struct {
	ResponseTimes           timestats.TimeStats
	StatusCodesDistribution map[string]int
	RequestEventTimes       map[string]timestats.TimeStats
	Records                 []struct {
		ResponseTime time.Duration
	}
	RequestFailures []struct {
		Reason string
	}
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

	times := make([]time.Duration, len(records))
	for i, rec := range records {
		agg.Records = append(agg.Records, struct{ ResponseTime time.Duration }{rec.Time})
		times[i] = rec.Time
		if rec.Error != "" {
			agg.RequestFailures = append(agg.RequestFailures, struct{ Reason string }{rec.Error})
		}
	}

	agg.ResponseTimes = timestats.Compute(times)

	agg.StatusCodesDistribution = computeStatusCodesDistribution(records)

	agg.RequestEventTimes = computeRequestEventTimes(records)

	return agg
}
