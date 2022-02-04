package dispatcher_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/benchttp/runner/dispatcher"
)

func TestNew(t *testing.T) {
	t.Run("panic if numWorker < 1", func(t *testing.T) {
		for _, numWorker := range []int{-1, 0} {
			func(numWorker int) {
				expMessage := fmt.Sprintf(
					"invalid numWorker value: must be > 1, got %d",
					numWorker,
				)

				defer func() {
					r := recover()
					if r == nil {
						t.Error("expected to panic but did not")
					}
					if r != expMessage {
						t.Errorf("unexpected panic message:\nexp %s\ngot %v", expMessage, r)
					}
				}()

				if d := dispatcher.New(numWorker); d != nil {
					t.Error("returned a non-nil Dispatcher")
				}
			}(numWorker)
		}
	})

	t.Run("return valid Dispatcher if numWorker > 0", func(t *testing.T) {
		if d := dispatcher.New(10); d == nil {
			t.Error("returned nil Dispatcher")
		}
	})
}

func TestDo(t *testing.T) {
	t.Run("stop when maxIter is reached", func(t *testing.T) {
		const (
			numWorker = 1
			maxIter   = 10
			expIter   = 10
		)

		gotIter := 0

		dispatcher.New(numWorker).Do(context.Background(), maxIter, func() { //nolint:errcheck
			gotIter++
		})

		if gotIter != expIter {
			t.Errorf("iterations: exp %d, got %d", expIter, gotIter)
		}
	})

	t.Run("stop on context timeout", func(t *testing.T) {
		const (
			timeout   = 100 * time.Millisecond
			interval  = timeout / 10
			numWorker = 1
			maxIter   = -1

			margin      = 25 * time.Millisecond // determined empirically
			maxDuration = timeout + margin
		)

		var (
			expIterMax = int(interval.Milliseconds()) + 1 // should not be reached
			gotIter    = 0
		)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		gotDuration := timeFunc(func() {
			dispatcher.New(numWorker).Do(ctx, maxIter, func() { //nolint:errcheck
				gotIter++
				time.Sleep(interval)
			})
		})

		if gotDuration > maxDuration {
			t.Errorf(
				"context timeout duration: exp < %dms, got %dms",
				maxDuration.Milliseconds(), gotDuration.Milliseconds(),
			)
		}

		if gotIter >= expIterMax {
			t.Errorf(
				"context timeout iterations: exp < %d, got %d",
				expIterMax, gotIter,
			)
		}
	})

	t.Run("stop on context cancel", func(t *testing.T) {
		const (
			timeout   = 100 * time.Millisecond
			interval  = timeout / 10
			numWorker = 1
			maxIter   = -1

			margin      = 25 * time.Millisecond // determined empirically
			maxDuration = timeout + margin
		)

		var (
			expIterMax = int(interval.Milliseconds()) + 1 // should not be reached
			gotIter    = 0
		)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(timeout)
			cancel()
		}()

		gotDuration := timeFunc(func() {
			dispatcher.New(numWorker).Do(ctx, maxIter, func() { //nolint:errcheck
				time.Sleep(interval)
			})
		})

		if gotDuration > maxDuration {
			t.Errorf(
				"context cancel duration: exp < %dms, got %dms",
				maxDuration.Milliseconds(), gotDuration.Milliseconds(),
			)
		}

		if gotIter >= expIterMax {
			t.Errorf(
				"context timeout iterations: exp < %d, got %d",
				expIterMax, gotIter,
			)
		}
	})

	t.Run("limit concurrent workers", func(t *testing.T) {
		const (
			interval  = 10 * time.Millisecond
			numWorker = 10
			maxIter   = 100

			// occasionnally we can have 1 extra concurrent goroutine,
			// we consider it an acceptable error margin
			margin             = 1
			expMaxNumGoroutine = numWorker + margin
		)

		var (
			mu               sync.Mutex
			baseNumGoroutine = runtime.NumGoroutine()
			gotNumGoroutines = make([]int, 0, maxIter)
		)

		dispatcher.New(numWorker).Do(context.Background(), maxIter, func() { //nolint:errcheck
			mu.Lock()
			gotNumGoroutines = append(gotNumGoroutines, runtime.NumGoroutine()-baseNumGoroutine)
			mu.Unlock()
			time.Sleep(interval)
		})

		for _, gotNumGoroutine := range gotNumGoroutines {
			if gotNumGoroutine > expMaxNumGoroutine {
				t.Errorf(
					"max concurrent workers: exp <= %d, got %d",
					expMaxNumGoroutine, gotNumGoroutine,
				)
			}
		}

		t.Log(gotNumGoroutines)
	})

	t.Run("dispatch concurrent workers correctly", func(t *testing.T) {
		const (
			numWorker = 3
			maxIter   = 12

			minIntervalBetweenGroups = 30 * time.Millisecond
			maxIntervalWithinGroup   = 10 * time.Millisecond
		)

		var (
			// elapsedTimes is a slice of durations corresponding to the
			// intervals between the call to semimpl.Do and each callback.
			elapsedTimes = make([]time.Duration, 0, maxIter)
			mu           sync.Mutex
		)

		start := time.Now()
		dispatcher.New(numWorker).Do(context.Background(), maxIter, func() { //nolint:errcheck
			mu.Lock()
			elapsedTimes = append(elapsedTimes, time.Since(start))
			mu.Unlock()
			time.Sleep(minIntervalBetweenGroups)
		})

		// check elapsedTimes slice is coherent, grouping its values
		// by expectedly similar durations, e.g.:
		// 12 iterations / 3 workers -> 4 groups of 3 similar durations.
		// With a callback duration of 30ms, we can expect such grouping:
		// [[0ms, 0ms, 0ms], [30ms, 30ms, 30ms], [60ms, 60ms, 60ms], [90ms, 90ms, 90ms]]
		// with a certain delta.
		// We check the resulting grouping against 2 rules:
		// 	1. durations within a same group must be close
		// 	2. max interval between two groups must be higher than the callback duration
		groups := groupby(elapsedTimes, numWorker)
		for groupIndex, group := range groups {
			// 1. check durations within each group are similar
			hi, lo := maxof(group), minof(group)
			if interval := hi - lo; interval > maxIntervalWithinGroup {
				t.Errorf(
					"unexpected interval in group: exp < %dms, got %dms",
					maxIntervalWithinGroup.Milliseconds(), interval.Milliseconds(),
				)
			}

			// check durations between distinct groups are spaced
			if groupIndex == len(groups)-1 {
				break
			}
			curr, next := minof(group), minof(groups[groupIndex+1])
			if interval := next - curr; interval < minIntervalBetweenGroups {
				t.Errorf(
					"unexpected interval between groups: exp > %dms, got %dms",
					minIntervalBetweenGroups.Milliseconds(), interval.Milliseconds(),
				)
			}
		}

		t.Log(elapsedTimes)
	})
}

func TestValidate(t *testing.T) {
	testcases := []struct {
		label     string
		exp       error
		numWorker int
		maxIter   int
		callback  func()
	}{
		{
			label:     "return error if maxIter == 0",
			exp:       errors.New("invalid value: maxIter: must be -1 or >= 1, got 0"),
			numWorker: 10,
			maxIter:   0,
			callback:  func() {},
		},
		{
			label:     "return error if maxIter == -2",
			exp:       errors.New("invalid value: maxIter: must be -1 or >= 1, got -2"),
			numWorker: 10,
			maxIter:   -2,
			callback:  func() {},
		},
		{
			label:     "return error if maxIter < numWorker",
			exp:       errors.New("invalid value: maxIter: must be >= numWorker (10), got 5"),
			numWorker: 10,
			maxIter:   5,
			callback:  func() {},
		},
		{
			label:     "return error if callback == nil",
			exp:       errors.New("invalid value: callback: must be non-nil"),
			numWorker: 10,
			maxIter:   20,
			callback:  nil,
		},
		{
			label:     "return nil if values are valid (maxIter == 1)",
			exp:       nil,
			numWorker: 1,
			maxIter:   1,
			callback:  func() {},
		},
		{
			label:     "return nil if values are valid (maxIter == -1)",
			exp:       nil,
			numWorker: 1,
			maxIter:   -1,
			callback:  func() {},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			got := dispatcher.New(tc.numWorker).Do(ctx, tc.maxIter, tc.callback)
			if got != nil && got.Error() != tc.exp.Error() {
				t.Errorf("unexpected error value:\nexp %v\ngot %v", tc.exp, got)
			}
		})
	}
}

// helpers

func groupby(src []time.Duration, by int) [][]time.Duration {
	numGroups := len(src) / by
	out := make([][]time.Duration, 0, numGroups)

	for i := 0; i < numGroups; i++ {
		lo := by * i
		hi := lo + by
		out = append(out, src[lo:hi])
	}

	return out
}

func minof(src []time.Duration) time.Duration {
	var min time.Duration
	for _, d := range src {
		if d < min || min == 0 {
			min = d
		}
	}
	return min
}

func maxof(src []time.Duration) time.Duration {
	var max time.Duration
	for _, d := range src {
		if d > max {
			max = d
		}
	}
	return max
}

func timeFunc(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}
