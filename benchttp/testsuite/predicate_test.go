package testsuite_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/benchttp/metrics"
	"github.com/benchttp/engine/benchttp/testsuite"
)

func TestPredicate(t *testing.T) {
	const (
		base = 100
		more = base + 1
		less = base - 1
	)

	testcases := []struct {
		Predicate  testsuite.Predicate
		PassValues []int
		FailValues []int
	}{
		{
			Predicate:  testsuite.EQ,
			PassValues: []int{base},
			FailValues: []int{more, less},
		},
		{
			Predicate:  testsuite.NEQ,
			PassValues: []int{less, more},
			FailValues: []int{base},
		},
		{
			Predicate:  testsuite.LT,
			PassValues: []int{more},
			FailValues: []int{base, less},
		},
		{
			Predicate:  testsuite.LTE,
			PassValues: []int{more, base},
			FailValues: []int{less},
		},
		{
			Predicate:  testsuite.GT,
			PassValues: []int{less},
			FailValues: []int{base, more},
		},
		{
			Predicate:  testsuite.GTE,
			PassValues: []int{less, base},
			FailValues: []int{more},
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.Predicate)+":pass", func(t *testing.T) {
			for _, passValue := range tc.PassValues {
				expectPredicatePass(t, tc.Predicate, base, passValue)
			}
		})
		t.Run(string(tc.Predicate+":fail"), func(t *testing.T) {
			for _, failValue := range tc.FailValues {
				expectPredicateFail(t, tc.Predicate, base, failValue)
			}
		})
	}
}

func expectPredicatePass(
	t *testing.T,
	p testsuite.Predicate,
	src, tar int,
) {
	t.Helper()
	expectPredicateResult(t, p, src, tar, true)
}

func expectPredicateFail(
	t *testing.T,
	p testsuite.Predicate,
	src, tar int,
) {
	t.Helper()
	expectPredicateResult(t, p, src, tar, false)
}

func expectPredicateResult(
	t *testing.T,
	p testsuite.Predicate,
	src, tar int,
	expPass bool,
) {
	t.Helper()

	agg := metrics.Aggregate{
		Records: make([]struct {
			ResponseTime time.Duration
		}, src),
	}

	result := testsuite.Run(agg, []testsuite.Case{{
		Predicate: p,
		Field:     "RequestCount",
		Target:    tar,
	}})

	if pass := result.Pass; pass != expPass {
		t.Errorf(
			"exp %v %d %d -> %v, got %v",
			p, src, tar, expPass, pass,
		)
	}
}
