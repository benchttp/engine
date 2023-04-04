package metrics

import (
	"fmt"
	"time"
)

// ComparisonResult is the result of a comparison.
type ComparisonResult int

const (
	// INF is the result of an inferiority check.
	INF ComparisonResult = -1
	// EQ is the result of an equality check.
	EQ ComparisonResult = 0
	// INF is the result of superiority check.
	SUP ComparisonResult = 1
)

// comapreMetrics compares the values of m and n,
// and returns the result from the point of view of n.
//
// It panics if m and n are not of the same type,
// or if their type is not handled.
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

// compareInts compares a and b and returns a ComparisonResult
// from the point of view of b.
func compareInts(a, b int) ComparisonResult {
	if b < a {
		return INF
	}
	if b > a {
		return SUP
	}
	return EQ
}

// compareInts compares a and b and returns a ComparisonResult
// from the point of view of b.
func compareDurations(a, b time.Duration) ComparisonResult {
	return compareInts(int(a), int(b))
}

// assertInts returns a, b as ints and true if a and b
// are both ints, else it returns 0, 0, false.
func assertInts(a, b Value) (x, y int, ok bool) {
	x, ok = a.(int)
	if !ok {
		return
	}
	y, ok = b.(int)
	return
}

// assertInts returns a, b as time.Durations and true if a and b
// are both time.Duration, else it returns 0, 0, false.
func assertDurations(a, b Value) (x, y time.Duration, ok bool) {
	x, ok = a.(time.Duration)
	if !ok {
		return
	}
	y, ok = b.(time.Duration)
	return
}
