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
		}},
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
		}},
}

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
			Median: 175}

		got := metrics.ComputeRequestEventTimes(validRecordsWithEvents)

		for event, _ := range got {
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

// approxEqual returns true if val is equal to target with a margin of error.
func approxEqualTime(val, target, margin time.Duration) bool {
	return val >= target-margin && val <= target+margin
}
