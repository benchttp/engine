package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

func computeRequestEventTimes(records []recorder.Record) (requestEventTimes map[string]timestats.TimeStats) {
	requestEventTimes = make(map[string]timestats.TimeStats, 0)

	allEventsWithTime := extractAllEventsWithTime(records)

	EachEventWithTimes := make(map[string][]time.Duration, 0)
	for _, event := range allEventsWithTime {
		EachEventWithTimes[event.Name] = append(EachEventWithTimes[event.Name], event.Time)
	}

	for eventName, times := range EachEventWithTimes {
		requestEventTimes[eventName] = timestats.Compute(times)
	}

	return requestEventTimes
}

// extractAllEventsWithTime gets all events in all records of a slice of records, with their time.
// It takes care of calculating the time of each event, as opposed to
// the time since the beginning of the request as it appears in the records.
func extractAllEventsWithTime(records []recorder.Record) (allEventsWithTime []recorder.Event) {
	for _, record := range records {
		for i, event := range record.Events {
			if i > 0 {
				event = recorder.Event{Name: event.Name, Time: event.Time - record.Events[i-1].Time}
			}
			allEventsWithTime = append(allEventsWithTime, event)
		}
	}
	return allEventsWithTime
}
