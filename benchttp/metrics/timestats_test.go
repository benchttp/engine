package metrics_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/benchttp/metrics"
)

func TestCompute(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := metrics.TimeStats{
			Min:       100,
			Max:       400,
			Mean:      230,
			Median:    200,
			StdDev:    110,
			Deciles:   []time.Duration{100, 100, 200, 200, 200, 300, 300, 400, 400, 400},
			Quartiles: []time.Duration{100, 200, 300, 400},
		}

		data := []time.Duration{100, 100, 200, 300, 400, 200, 100, 200, 300, 400, 100, 100, 200, 300, 400, 200, 100, 200, 300, 400}

		got := metrics.NewTimeStats(data)

		for _, stat := range []struct {
			name string
			want time.Duration
			got  time.Duration
		}{
			{"min", want.Min, got.Min},
			{"max", want.Max, got.Max},
			{"mean", want.Mean, got.Mean},
			{"median", want.Median, got.Median},
			{"stdDev", want.StdDev, got.StdDev},
		} {
			if !approxEqualTime(stat.got, stat.want, 1) {
				t.Errorf("%s: want %d, got %d", stat.name, stat.want, stat.got)
			}
		}

		if len(got.Deciles) != len(want.Deciles) {
			t.Fatalf("deciles: want %d deciles, got %d", len(want.Deciles), len(got.Deciles))
		}

		for i := range got.Deciles {
			if got.Deciles[i] != want.Deciles[i] {
				t.Errorf("decile %d: want %d, got %d", i+1, want.Deciles[i], got.Deciles[i])
			}
		}

		if len(got.Quartiles) != len(want.Quartiles) {
			t.Fatalf("quartiles: want %d quartiles, got %d", len(want.Deciles), len(got.Quartiles))
		}

		for i := range got.Quartiles {
			if got.Quartiles[i] != want.Quartiles[i] {
				t.Errorf("decile %d: want %d, got %d", i+1, want.Quartiles[i], got.Quartiles[i])
			}
		}
	})

	t.Run("few values", func(t *testing.T) {
		data := []time.Duration{100, 300}
		got := metrics.NewTimeStats(data)

		if got.Deciles != nil {
			t.Errorf("deciles: want nil, got %v", got.Deciles)
		}

		if got.Quartiles != nil {
			t.Errorf("quartiles: want nil, got %v", got.Quartiles)
		}

		if got.Median != 200 {
			t.Errorf("median: want 200ns, got %v", got.Median)
		}
	})
}
