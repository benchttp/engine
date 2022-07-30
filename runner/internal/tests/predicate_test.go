package tests_test

import (
	"testing"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/tests"
)

func TestPredicate(t *testing.T) {
	const (
		metric    = 100
		metricInc = metric + 1
		metricDec = metric - 1
	)

	testcases := []struct {
		Predicate  tests.Predicate
		PassValues []int
		FailValues []int
	}{
		{
			Predicate:  tests.EQ,
			PassValues: []int{metric},
			FailValues: []int{metricDec, metricInc},
		},
		{
			Predicate:  tests.NEQ,
			PassValues: []int{metricInc, metricDec},
			FailValues: []int{metric},
		},
		{
			Predicate:  tests.LT,
			PassValues: []int{metricDec},
			FailValues: []int{metric, metricInc},
		},
		{
			Predicate:  tests.LTE,
			PassValues: []int{metricDec, metric},
			FailValues: []int{metricInc},
		},
		{
			Predicate:  tests.GT,
			PassValues: []int{metricInc},
			FailValues: []int{metric, metricDec},
		},
		{
			Predicate:  tests.GTE,
			PassValues: []int{metricInc, metric},
			FailValues: []int{metricDec},
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.Predicate)+":pass", func(t *testing.T) {
			for _, passValue := range tc.PassValues {
				expectPredicatePass(t, tc.Predicate, metric, passValue)
			}
		})
		t.Run(string(tc.Predicate+":fail"), func(t *testing.T) {
			for _, failValue := range tc.FailValues {
				expectPredicateFail(t, tc.Predicate, metric, failValue)
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

	cases := []tests.Case{{
		Predicate: p,
		Source:    metrics.RequestCount,
		Target:    metrics.Value(tar),
	}}

	result := tests.Run(agg, cases)

	if pass := result.Pass; pass != expPass {
		t.Errorf(
			"exp %v %d %d -> %v, got %v",
			p, src, tar, expPass, pass,
		)
	}
}
