package metrics

import (
	"time"

	"github.com/benchttp/engine/internal/timestats"
	"github.com/benchttp/engine/runner/internal/recorder"
)

func computeRequestEventTimes(records []recorder.Record) map[string]timestats.TimeStats {
	events := getFlatRelativeTimeEvents(records)

	timesByEvent := map[string][]time.Duration{}

	for _, e := range events {
		timesByEvent[e.Name] = append(timesByEvent[e.Name], e.Time)
	}

	statsByEvent := map[string]timestats.TimeStats{}

	for e, times := range timesByEvent {
		statsByEvent[e] = timestats.New(times)
	}

	return statsByEvent
}

func getFlatRelativeTimeEvents(records []recorder.Record) []recorder.Event {
	events := []recorder.Event{}
	for _, record := range records {
		events = append(events, record.RelativeTimeEvents()...)
	}
	return events
}
