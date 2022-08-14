package timestats_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/timestats"
)

var validTimes = []time.Duration{100, 100, 200, 300, 400, 200, 100, 200, 300, 400, 100, 100, 200, 300, 400, 200, 100, 200, 300, 400}

func TestCompute(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := timestats.TimeStats{
			Min:     100,
			Max:     400,
			Avg:     230,
			Median:  200,
			StdDev:  110,
			Deciles: [10]time.Duration{100, 100, 200, 200, 200, 300, 300, 400, 400, 400},
		}

		got := timestats.Compute(validTimes)

		for _, stat := range []struct {
			name string
			want time.Duration
			got  time.Duration
		}{
			{"min", want.Min, got.Min},
			{"max", want.Max, got.Max},
			{"avg", want.Avg, got.Avg},
			{"median", want.Median, got.Median},
			{"stdDev", want.StdDev, got.StdDev},
		} {
			if !approxEqualTime(stat.got, stat.want, 1) {
				t.Errorf("%s: want %d, got %d", stat.name, stat.want, stat.got)
			}
		}

		for i := range got.Deciles {
			if got.Deciles[i] != want.Deciles[i] {
				t.Errorf("decile %d: want %d, got %d", i+1, want.Deciles[i], got.Deciles[i])
			}
		}
	})
}

// approxEqual returns true if val is equal to target with a margin of error.
func approxEqualTime(val, target, margin time.Duration) bool {
	return val >= target-margin && val <= target+margin
}
