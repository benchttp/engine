package tests_test

import (
	"testing"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/tests"
)

func TestPredicate(t *testing.T) {
	const (
		source = 100
		same   = source
		more   = source + 1
		less   = source - 1
	)

	testcases := []struct {
		Predicate  tests.Predicate
		PassValues []int
		FailValues []int
	}{
		{
			Predicate:  tests.EQ,
			PassValues: []int{same},
			FailValues: []int{more, less},
		},
		{
			Predicate:  tests.NEQ,
			PassValues: []int{less, more},
			FailValues: []int{same},
		},
		{
			Predicate:  tests.LT,
			PassValues: []int{more},
			FailValues: []int{same, less},
		},
		{
			Predicate:  tests.LTE,
			PassValues: []int{more, same},
			FailValues: []int{less},
		},
		{
			Predicate:  tests.GT,
			PassValues: []int{less},
			FailValues: []int{same, more},
		},
		{
			Predicate:  tests.GTE,
			PassValues: []int{less, same},
			FailValues: []int{more},
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.Predicate)+":pass", func(t *testing.T) {
			for _, passValue := range tc.PassValues {
				expectPredicatePass(t, tc.Predicate, source, passValue)
			}
		})
		t.Run(string(tc.Predicate+":fail"), func(t *testing.T) {
			for _, failValue := range tc.FailValues {
				expectPredicateFail(t, tc.Predicate, source, failValue)
			}
		})
	}
}

func expectPredicatePass(
	t *testing.T,
	p tests.Predicate,
	src, tar int,
) {
	t.Helper()
	expectPredicateResult(t, p, src, tar, true)
}

func expectPredicateFail(
	t *testing.T,
	p tests.Predicate,
	src, tar int,
) {
	t.Helper()
	expectPredicateResult(t, p, src, tar, false)
}

func expectPredicateResult(
	t *testing.T,
	p tests.Predicate,
	src, tar int,
	expPass bool,
) {
	t.Helper()

	agg := metrics.Aggregate{
		TotalCount: src,
	}

	result := tests.Run(agg, []tests.Case{{
		Predicate: p,
		Source:    metrics.RequestCount,
		Target:    tar,
	}})

	if pass := result.Pass; pass != expPass {
		t.Errorf(
			"exp %v %d %d -> %v, got %v",
			p, src, tar, expPass, pass,
		)
	}
}
