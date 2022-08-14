package timestats

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type TimeStats struct {
	Min, Max, Avg, Median, StdDev time.Duration
	Quartiles                     [4]time.Duration
	Deciles                       [10]time.Duration
}

func Compute(times []time.Duration) (timeStats TimeStats) {
	n := len(times)
	if n == 0 {
		return
	}

	var sum, avg time.Duration
	comparableDurations := make(comparableDurations, n)
	for i, time := range times {
		comparableDurations[i] = time
		sum += time
	}

	sort.Sort(comparableDurations)
	avg = sum / time.Duration(n)

	fmt.Println(comparableDurations)

	return TimeStats{
		Min:       comparableDurations[0],
		Max:       comparableDurations[len(comparableDurations)-1],
		Avg:       avg,
		Median:    calculateMedian(comparableDurations),
		StdDev:    time.Duration(calculateStdDev(times, avg)),
		Quartiles: calculateQuartiles(comparableDurations),
		Deciles:   calculateDeciles(comparableDurations),
	}
}

func calculateMedian(sorted []time.Duration) time.Duration {
	n := len(sorted)
	if n%2 != 0 {
		return sorted[(n/2)-1]
	}
	return (sorted[(n/2)-1] + sorted[(n/2)]) / 2
}

func calculateStdDev(values []time.Duration, avg time.Duration) time.Duration {
	n := len(values)
	sum := time.Duration(0)
	for _, v := range values {
		dev := v - avg
		sum += dev * dev
	}
	return time.Duration(math.Sqrt(float64(sum / time.Duration(n))))
}

func calculateDeciles(sorted []time.Duration) [10]time.Duration {
	const numDecile = 10
	return *(*[10]time.Duration)(calculateQuantiles(sorted, numDecile))
}

func calculateQuartiles(sorted []time.Duration) [4]time.Duration {
	const numQuartile = 4
	return *(*[4]time.Duration)(calculateQuantiles(sorted, numQuartile))
}

func calculateQuantiles(sorted []time.Duration, numQuantile int) []time.Duration {
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

type comparableDurations []time.Duration

func (s comparableDurations) Len() int {
	return len(s)
}

func (s comparableDurations) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s comparableDurations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
