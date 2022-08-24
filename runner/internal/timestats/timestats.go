package timestats

import (
	"math"
	"sort"
	"time"
)

type TimeStats struct {
	Min, Max, Mean, Median, StdDev time.Duration
	Quartiles                      [4]time.Duration
	Deciles                        [10]time.Duration
}

func Compute(times []time.Duration) TimeStats {
	l := len(times)
	if l == 0 {
		return TimeStats{}
	}

	// Measures computing functions works on sorted data.
	// Sort once and compute upon the result.
	sort.Sort(byFastest(times))

	// Reused statistics measures.
	sum := computeSum(times)
	mean := computeMean(sum, l)

	return TimeStats{
		Min:       times[0],
		Max:       times[len(times)-1],
		Mean:      mean,
		Median:    computeMedian(times),
		StdDev:    computeStdDev(times, mean),
		Quartiles: computeQuartiles(times),
		Deciles:   computeDeciles(times),
	}
}

func computeSum(values []time.Duration) time.Duration {
	var sum time.Duration
	for _, time := range values {
		sum += time
	}
	return sum
}

func computeMean(sum time.Duration, length int) time.Duration {
	return sum / time.Duration(length)
}

func computeMedian(sorted []time.Duration) time.Duration {
	l := len(sorted)
	if l%2 != 0 {
		return sorted[l/2]
	}
	return computeMean(sorted[l/2-1]+sorted[l/2+1], 2)
}

func computeStdDev(values []time.Duration, mean time.Duration) time.Duration {
	sum := time.Duration(0)
	for _, v := range values {
		dev := v - mean
		sum += dev * dev
	}
	return time.Duration(math.Sqrt(float64(sum / time.Duration(len(values)))))
}

func computeDeciles(sorted []time.Duration) [10]time.Duration {
	const numDecile = 10
	if len(sorted) < numDecile {
		return [10]time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	return *(*[10]time.Duration)(computeQuantiles(sorted, numDecile))
}

func computeQuartiles(sorted []time.Duration) [4]time.Duration {
	const numQuartile = 4
	if len(sorted) < numQuartile {
		return [4]time.Duration{0, 0, 0, 0}
	}
	return *(*[4]time.Duration)(computeQuantiles(sorted, numQuartile))
}

func computeQuantiles(sorted []time.Duration, numQuantile int) []time.Duration {
	numValues := len(sorted)
	step := (numValues + 1) / numQuantile

	quantiles := make([]time.Duration, numQuantile)
	for i := 0; i < numQuantile; i++ {
		qtlIndex := (i + 1) * step
		maxIndex := numValues - 1
		if qtlIndex > maxIndex {
			qtlIndex = maxIndex
		}
		quantiles[i] = sorted[qtlIndex]
	}
	return quantiles
}
