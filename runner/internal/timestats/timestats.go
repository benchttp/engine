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

func New(times []time.Duration) TimeStats {
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
	if n%2 != 0 {
		return sorted[n/2]
	}
	return computeMean(sorted[n/2-1]+sorted[n/2+1], 2)
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

func computeDeciles(sorted []time.Duration) [10]time.Duration {
	const nDeciles = 10
	if len(sorted) < nDeciles {
		return [10]time.Duration{}
	}
	return *(*[10]time.Duration)(computeQuantiles(sorted, nDeciles))
}

func computeQuartiles(sorted []time.Duration) [4]time.Duration {
	const nQuartiles = 4
	if len(sorted) < nQuartiles {
		return [4]time.Duration{}
	}
	return *(*[4]time.Duration)(computeQuantiles(sorted, nQuartiles))
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
