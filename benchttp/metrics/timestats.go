package metrics

import (
	"math"
	"sort"
	"time"
)

const (
	numDecile   = 10
	numQuartile = 4
)

type TimeStats struct {
	Min, Max, Mean, Median, StdDev time.Duration
	Quartiles                      []time.Duration
	Deciles                        []time.Duration
}

func NewTimeStats(times []time.Duration) TimeStats {
	n := len(times)
	if n == 0 {
		return TimeStats{}
	}

	// Measures computing functions works on sorted data.
	// Sort once and compute upon the result.
	sort.Sort(byFastest(times))

	// Reused statistics measures.
	sum := computeSum(times)
	mean := computeMean(sum, n)

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
	n := len(sorted)
	mid := n / 2
	if odd := n&1 == 1; odd {
		return sorted[mid]
	}
	himid := mid
	lomid := himid - 1
	return computeMean(sorted[lomid]+sorted[himid], 2)
}

func computeStdDev(values []time.Duration, mean time.Duration) time.Duration {
	sum := time.Duration(0)
	for _, v := range values {
		dev := v - mean
		sum += dev * dev
	}
	n := len(values)
	return time.Duration(math.Sqrt(float64(sum / time.Duration(n))))
}

func computeDeciles(sorted []time.Duration) []time.Duration {
	if len(sorted) < numDecile {
		return nil
	}
	return computeQuantiles(sorted, numDecile)
}

func computeQuartiles(sorted []time.Duration) []time.Duration {
	if len(sorted) < numQuartile {
		return nil
	}
	return computeQuantiles(sorted, numQuartile)
}

func computeQuantiles(sorted []time.Duration, nQuantiles int) []time.Duration {
	n := len(sorted)
	step := (n + 1) / nQuantiles

	quantiles := make([]time.Duration, nQuantiles)
	for i := 0; i < nQuantiles; i++ {
		qtlIndex := (i + 1) * step
		maxIndex := n - 1
		if qtlIndex > maxIndex {
			qtlIndex = maxIndex
		}
		quantiles[i] = sorted[qtlIndex]
	}
	return quantiles
}
