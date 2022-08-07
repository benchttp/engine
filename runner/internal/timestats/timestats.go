package timestats

import (
	"fmt"
	"sort"
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/montanaflynn/stats"
)

type TimeStats struct {
	Min, Max, Avg, Median, StdDev time.Duration
	Deciles                       map[int]time.Duration
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
	stdDev, stdDevErrs := pipe("avg", errs)(stats.StandardDeviation(times))
	if avgErrs != nil {
		errs = append(errs, stdDevErrs...)
	}

	deciles := map[int]float64{1: 10, 2: 20, 3: 30, 4: 40, 5: 50, 6: 60, 7: 70, 8: 80, 9: 90}

	keys := make([]int, 0)
	for k := range deciles {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	decilesErrs := []error{}

	for _, k := range keys {
		var decileErrs []error = make([]error, 1)
		float64v := float64(deciles[k])
		n := fmt.Sprintf("%s decile", ordinal(k))
		deciles[k], decileErrs = pipe(n, errs)(stats.Percentile(times, float64v))
		if len(decileErrs) > 0 {
			errs = append(decilesErrs, decileErrs...)
		}
	}

	if len(decilesErrs) > 0 {
		errs = append(errs, decilesErrs...)
	}

	if len(errs) > 0 {
		return timeStats, errs
	}

	timeStats = convertTimeStatsBackToTimeDuration(min, max, avg, median, stdDev, deciles)

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

func convertTimeStatsBackToTimeDuration(min float64, max float64, avg float64, median float64, stdDev float64, deciles map[int]float64) (timeStats TimeStats) {
	timeStats.Min = time.Duration(min)
	timeStats.Max = time.Duration(max)
	timeStats.Avg = time.Duration(avg)
	timeStats.Median = time.Duration(median)
	timeStats.StdDev = time.Duration(stdDev)

	fmt.Println("TEST1")

	timeStats.Deciles = make(map[int]time.Duration, 9)

	for i, p := range deciles {
		timeStats.Deciles[i] = time.Duration(p)
	}

	fmt.Println("TEST2")

	fmt.Println(timeStats.Deciles)

	return timeStats
}

// ordinal return x ordinal format.
//	ordinal(3) == "3rd"
func ordinal(x int) string {
	suffix := "th"
	switch x % 10 {
	case 1:
		if x%100 != 11 {
			suffix = "st"
		}
	case 2:
		if x%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if x%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", x, suffix)
}
