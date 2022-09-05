package metrics_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/metrics/timestats"
	"github.com/benchttp/engine/runner/internal/recorder"
)

func TestNewAggregate(t *testing.T) {
	// Test for "response times stats" is delegated to timestats.New because
	// metrics.NewAggregate does not have any specific behavior aound it.

	t.Run("events times stats", func(t *testing.T) {
		eventsStub := func(t1, t2 time.Duration) []recorder.Event {
			return []recorder.Event{{Name: "1", Time: t1}, {Name: "2", Time: t2}}
		}

		input := []recorder.Record{
			{Events: eventsStub(100, 200)},
			{Events: eventsStub(200, 200)},
			{Events: eventsStub(300, 400)},
			{Events: eventsStub(400, 500)},
		}

		want := map[string]timestats.TimeStats{
			"1": {Min: 100, Max: 400, Mean: 250, Median: 300},
			"2": {Min: 200, Max: 500, Mean: 325, Median: 350},
		}

		got := metrics.NewAggregate(input).RequestEventTimes

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
					t.Errorf("RequestEventTimes: %s: %s: want %d, got %d",
						event, stat.name, stat.want, stat.got)
				}
			}
		}
	})

	t.Run("status codes stats", func(t *testing.T) {
		input := []recorder.Record{
			{Code: 200}, {Code: 200}, {Code: 200}, {Code: 400}, {Code: 400}, {Code: 500},
		}

		want := map[int]int{200: 3, 400: 2, 500: 1}

		got := metrics.NewAggregate(input).StatusCodesDistribution

		if reflect.ValueOf(got).IsZero() {
			t.Error("want stats output to be non-zero value, got zero value")
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("StatusCodesDistribution: want %v, got %v", want, got)
		}
	})

	t.Run("records", func(t *testing.T) {
		input := []recorder.Record{
			{Time: 100}, {Time: 50}, {Time: 100}, {Time: 200}, {Time: 150},
		}

		want := []struct{ ResponseTime time.Duration }{{100}, {50}, {100}, {200}, {150}}

		got := metrics.NewAggregate(input).Records

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Records: want %v, got %v", want, got)
		}
	})

	t.Run("request failures", func(t *testing.T) {
		input := []recorder.Record{
			{Error: "something"}, {Error: "went"}, {Error: "wrong"},
		}

		want := []struct{ Reason string }{{"something"}, {"went"}, {"wrong"}}

		got := metrics.NewAggregate(input).RequestFailures

		if !reflect.DeepEqual(got, want) {
			t.Errorf("RequestFailures: want %v, got %v", want, got)
		}
	})
}

// approxEqual returns true if val is equal to target with a margin of error.
func approxEqualTime(val, target, margin time.Duration) bool {
	return val >= target-margin && val <= target+margin
}
