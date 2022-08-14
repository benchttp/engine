package metrics_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

var validRecordsWithEvents = []recorder.Record{
	{
		Events: []recorder.Event{
			{Name: "DNSDone", Time: 100},
			{Name: "ConnectDone", Time: 250},
			{Name: "DNSDone", Time: 350},
			{Name: "DNSDone", Time: 450},
			{Name: "DNSDone", Time: 550},
			{Name: "DNSDone", Time: 650},
			{Name: "DNSDone", Time: 750},
			{Name: "DNSDone", Time: 850},
			{Name: "DNSDone", Time: 950},
			{Name: "DNSDone", Time: 1050},
		},
	},
	{
		Events: []recorder.Event{
			{Name: "DNSDone", Time: 100},
			{Name: "ConnectDone", Time: 250},
			{Name: "ConnectDone", Time: 400},
			{Name: "ConnectDone", Time: 550},
			{Name: "ConnectDone", Time: 700},
			{Name: "ConnectDone", Time: 900},
			{Name: "ConnectDone", Time: 1100},
			{Name: "ConnectDone", Time: 1300},
			{Name: "ConnectDone", Time: 1500},
			{Name: "ConnectDone", Time: 1700},
		},
	},
}

// We only check some metrics to be confident that the time.Duration of the different
// events are well handled.
// All timestats are checked in timestats_test.go.
func TestComputeRequestEventTimes(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := map[string]timestats.TimeStats{}
		want["DNSDone"] = timestats.TimeStats{
			Min:    100,
			Max:    100,
			Avg:    100,
			Median: 100,
		}
		want["ConnectDone"] = timestats.TimeStats{
			Min:    150,
			Max:    200,
			Avg:    175,
			Median: 175,
		}

		got := metrics.ComputeRequestEventTimes(validRecordsWithEvents)

		for event := range got {
			for _, stat := range []struct {
				name string
				want time.Duration
				got  time.Duration
			}{
				{"min", want[event].Min, got[event].Min},
				{"max", want[event].Max, got[event].Max},
				{"avg", want[event].Avg, got[event].Avg},
				{"median", want[event].Median, got[event].Median},
			} {
				if !approxEqualTime(stat.got, stat.want, 1) {
					t.Errorf("%s: %s: want %d, got %d", event, stat.name, stat.want, stat.got)
				}
			}
		}
	})
}

func TestComputeRequestEventsDistribution(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := map[string]int{"DNSDone": 10, "ConnectDone": 10}

		got, errs := metrics.ComputeRequestEventsDistribution(validRecordsWithEvents)

		if len(errs) > 0 {
			t.Errorf("expected nil error, got %v", errs)
		}

		for event, count := range got {
			if got[event] != want[event] {
				t.Errorf("event %s: expected count %v, got %v", event, want, count)
			}
		}
	})

	t.Run("invalid event name provided", func(t *testing.T) {
		invalidRecordsWithEvents := []recorder.Record{
			{
				Events: []recorder.Event{
					{Name: "not_a_valid_event_name", Time: 100},
				},
			},
		}
		want := "not_a_valid_event_name is not a valid event name"

		_, errs := metrics.ComputeRequestEventsDistribution(invalidRecordsWithEvents)

		if len(errs) == 0 {
			t.Fatalf("want error, got none")
		}

		if errs[0].Error() != want {
			t.Errorf("did not get expected error: want %v, got %v", want, errs)
		}
	})
}

// approxEqual returns true if val is equal to target with a margin of error.
func approxEqualTime(val, target, margin time.Duration) bool {
	return val >= target-margin && val <= target+margin
}
