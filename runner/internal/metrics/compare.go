package metrics

import (
	"fmt"
	"time"
)

type ComparisonResult int

const (
	INF ComparisonResult = -1
	EQ  ComparisonResult = 0
	SUP ComparisonResult = 1
)

func compareMetrics(m, n Metric) ComparisonResult {
	a, b := m.Value, n.Value
	if a, b, isDuration := assertDurations(a, b); isDuration {
		return compareDurations(a, b)
	}
	if a, b, isInt := assertInts(a, b); isInt {
		return compareInts(a, b)
	}
	panic(fmt.Sprintf(
		"metrics: unhandled comparison: %v (%T) and %v (%T)",
		a, a, b, b,
	))
}

func compareInts(a, b int) ComparisonResult {
	if b < a {
		return INF
	}
	if b > a {
		return SUP
	}
	return EQ
}

func compareDurations(a, b time.Duration) ComparisonResult {
	return compareInts(int(a), int(b))
}

func assertInts(a, b Value) (x, y int, ok bool) {
	x, ok = a.(int)
	if !ok {
		return
	}
	y, ok = b.(int)
	return
}

func assertDurations(a, b Value) (x, y time.Duration, ok bool) {
	x, ok = a.(time.Duration)
	if !ok {
		return
	}
	y, ok = b.(time.Duration)
	return
}
