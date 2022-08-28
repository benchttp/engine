package metrics

import (
	"time"

	"github.com/benchttp/engine/internal/timestats"
	"github.com/benchttp/engine/runner/internal/recorder"
)

// Aggregate is an aggregate of metrics computed from
// a slice of recorder.Record.
type Aggregate struct {
	// ResponseTimes is the common statistics computed from a
	// slice recorder.Record. It offers statistics about the
	// recorder.Record.Time of the records.
	ResponseTimes timestats.TimeStats
	// StatusCodesDistribution maps each status code received to
	// its number of occurrence.
	StatusCodesDistribution map[string]int
	// RequestEventTimes is the common statistics computed from
	// the combination of all recorder.Record.Events of a slice
	// of recorder.Record. It offers statistics about the
	// recorder.Events.Time of the records.
	RequestEventTimes map[string]timestats.TimeStats
	// Records lists each response time received during the run.
	// It offers raw informarion.
	Records []struct {
		ResponseTime time.Duration
	}
	// Records lists each request error received during the run.
	// It offers raw informarion.
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

// NewAggregate computes and aggregates metrics from the given records.
func NewAggregate(records []recorder.Record) (agg Aggregate) {
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

	agg.ResponseTimes = timestats.New(times)

	agg.StatusCodesDistribution = computeStatusCodesDistribution(records)

	agg.RequestEventTimes = computeRequestEventTimes(records)

	return agg
}
