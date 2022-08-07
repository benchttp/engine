package timestats

import (
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/montanaflynn/stats"
)

type TimeStats struct {
	Min, Max, Avg, Median time.Duration
}

func Compute(records []recorder.Record) (timeStats TimeStats, errs []error) {
	if len(records) == 0 {
		return timeStats, append(errs, ErrEmptySlice)
	}

	times := getFloat64Times(records)

	min, minErrs := pipe("min", errs)(stats.Min(times))
	if minErrs != nil {
		errs = append(errs, minErrs...)
	}
	max, maxErrs := pipe("max", errs)(stats.Max(times))
	if maxErrs != nil {
		errs = append(errs, maxErrs...)
	}
	avg, avgErrs := pipe("avg", errs)(stats.Mean(times))
	if avgErrs != nil {
		errs = append(errs, avgErrs...)
	}
	median, medianErrs := pipe("avg", errs)(stats.Median(times))
	if avgErrs != nil {
		errs = append(errs, medianErrs...)
	}

	if errs != nil {
		return timeStats, errs
	}

	timeStats = convertTimeStatsBackToTimeDuration(min, max, avg, median)

	return timeStats, nil
}

func pipe(name string, errs []error) func(float64, error) (float64, []error) {
	return func(stat float64, err error) (float64, []error) {
		if err != nil {
			errs = append(errs, ComputeError(name))
		}
		return stat, errs
	}
}

// github.com/montanaflynn/stats needs to be provided with float64, not time.Duration
func getFloat64Times(records []recorder.Record) []float64 {
	float64Times := make([]float64, len(records))
	for i, v := range records {
		float64Times[i] = float64(v.Time)
	}
	return float64Times
}

func convertTimeStatsBackToTimeDuration(min float64, max float64, avg float64, median float64) (timeStats TimeStats) {
	timeStats.Min = time.Duration(min)
	timeStats.Max = time.Duration(max)
	timeStats.Avg = time.Duration(avg)
	timeStats.Median = time.Duration(median)

	return timeStats
}
