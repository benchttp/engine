package timestats

import (
	"fmt"
	"time"

	"github.com/montanaflynn/stats"
)

type TimeStats struct {
	Min, Max, Avg time.Duration
}

func Compute(times []time.Duration) (timeStats TimeStats, errs []string) {
	float64Times := getFloat64Times(times)

	min, errs := pipe("min", errs)(stats.Min(float64Times))
	max, errs := pipe("max", errs)(stats.Max(float64Times))
	avg, errs := pipe("mean", errs)(stats.Mean(float64Times))

	if errs != nil {
		return timeStats, errs
	}

	timeStats = convertTimeStatsBackToTimeDuration(min, max, avg)

	return timeStats, errs
}

func getFloat64Times(times []time.Duration) []float64 {
	float64Times := make([]float64, len(times))
	for i, v := range times {
		float64Times[i] = float64(v)
	}
	return float64Times
}

func pipe(name string, errs []string) func(float64, error) (float64, []string) {
	return func(stat float64, err error) (float64, []string) {
		if err != nil {
			errs = append(errs, fmt.Sprintf("computing %s: %s", name, err))
		}
		return stat, errs
	}
}

func convertTimeStatsBackToTimeDuration(min float64, max float64, avg float64) (timeStats TimeStats) {
	timeStats.Min = time.Duration(min)
	timeStats.Max = time.Duration(max)
	timeStats.Avg = time.Duration(avg)

	return timeStats
}
