package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

func computeRequestEventTimes(records []recorder.Record) map[string]timestats.TimeStats {
	events := getUnnestedRelativeTimeEvents(records)

	timesByEvent := map[string][]time.Duration{}

	for _, e := range events {
		timesByEvent[e.Name] = append(timesByEvent[e.Name], e.Time)
	}

	statsByEvent := map[string]timestats.TimeStats{}

	for e, times := range timesByEvent {
		statsByEvent[e] = timestats.Compute(times)
	}

	return statsByEvent
}

func getUnnestedRelativeTimeEvents(records []recorder.Record) []recorder.Event {
	events := []recorder.Event{}
	for _, record := range records {
		events = append(events, record.RelativeTimeEvents()...)
	}
	return events
}
