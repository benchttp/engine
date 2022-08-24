package metrics

import (
	"math"
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

func computeRequestEventTimes(records []recorder.Record) map[string]timestats.TimeStats {
	events := getUnnestedDiffedEvents(records)

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

func getUnnestedDiffedEvents(records []recorder.Record) []recorder.Event {
	events := []recorder.Event{}
	for _, record := range records {
		events = append(events, diffEventsTimes(record.Events)...)
	}
	return events
}

// TODO It is weird that we create an entirely new Event slice
// with updated Time field. Think about making this a method of
// recorder.Record?

func diffEventsTimes(events []recorder.Event) []recorder.Event {
	diffed := make([]recorder.Event, len(events))
	for i, event := range events {
		switch i {
		case 0:
			diffed[i] = event
		default:
			diffed[i] = recorder.Event{Name: event.Name, Time: diff(event.Time, events[i-1].Time)}
		}
	}
	return diffed
}

// diff returns the time.Duration difference between a and b.
// a and b need not to be passed in specific order. The difference
// is expressed in absolute value.
func diff(a, b time.Duration) time.Duration {
	d := a - b
	switch {
	case d >= 0:
		return d
	case d == math.MinInt64:
		return math.MaxInt64
	default:
		return -d
	}
}
