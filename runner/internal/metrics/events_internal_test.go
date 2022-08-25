package metrics

import (
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

func TestComputeRequestEventTimes(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		got := computeRequestEventTimes([]recorder.Record{
			{
				Events: []recorder.Event{
					{Name: "1", Time: 100},
					{Name: "2", Time: 200}, // diff is 100
				},
			},
			{
				Events: []recorder.Event{
					{Name: "1", Time: 200},
					{Name: "2", Time: 200}, // diff is 0
				},
			},
			{
				Events: []recorder.Event{
					{Name: "1", Time: 300},
					{Name: "2", Time: 400}, // diff is 100
				},
			},
			{
				Events: []recorder.Event{
					{Name: "1", Time: 400},
					{Name: "2", Time: 500}, // diff is 100
				},
			},
		})

		want := map[string]timestats.TimeStats{
			"1": {
				Min:    100,
				Max:    400,
				Mean:   250,
				Median: 300,
			},
			"2": {
				Min:    0,
				Max:    100,
				Mean:   75,
				Median: 100,
			},
		}

		for event := range got {
			for _, stat := range []struct {
				name string
				want time.Duration
				got  time.Duration
			}{
				{"min", want[event].Min, got[event].Min},
				{"max", want[event].Max, got[event].Max},
				{"mean", want[event].Mean, got[event].Mean},
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
