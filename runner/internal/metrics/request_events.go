package metrics

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

func ComputeRequestEventTimes(records []recorder.Record) (requestEventTimes map[string]timestats.TimeStats) {
	requestEventTimes = make(map[string]timestats.TimeStats, 0)

	var allEvents []recorder.Event
	for _, record := range records {
		for i, event := range record.Events {
			if i > 0 {
				event = recorder.Event{Name: event.Name, Time: event.Time - record.Events[i-1].Time}
			}
			allEvents = append(allEvents, event)
		}
	}

	EachEventWithTimes := make(map[string][]time.Duration, 0)
	for _, event := range allEvents {
		EachEventWithTimes[event.Name] = append(EachEventWithTimes[event.Name], event.Time)
	}

	for eventName, times := range EachEventWithTimes {
		requestEventTimes[eventName] = timestats.Compute(times)
	}

	return requestEventTimes
}

func ComputeRequestEventsDistribution(records []recorder.Record) (requestEventsDistribution map[string]int, errs []error) {
	requestEventsDistribution = map[string]int{
		"DNSDone": 0, "ConnectDone": 0, "TLSHandshakeDone": 0, "WroteHeaders": 0, "WroteRequest": 0, "GotFirstResponseByte": 0, "PutIdleConn": 0,
	}

	var allEvents []recorder.Event
	for _, record := range records {
		allEvents = append(allEvents, record.Events...)
	}

	for _, event := range allEvents {
		_, isPresent := requestEventsDistribution[event.Name]
		if isPresent {
			requestEventsDistribution[event.Name]++
		} else {
			errs = append(errs, RequestEventsDistributionComputeErr(event.Name))
		}
	}

	if len(errs) > 0 {
		return requestEventsDistribution, errs
	}

	return requestEventsDistribution, nil
}
