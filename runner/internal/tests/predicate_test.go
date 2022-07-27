package tests_test

import (
	"testing"

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
		PassValues []tests.Value
		FailValues []tests.Value
	}{
		{
			Predicate:  tests.EQ,
			PassValues: []tests.Value{metric},
			FailValues: []tests.Value{metricDec, metricInc},
		},
		{
			Predicate:  tests.NEQ,
			PassValues: []tests.Value{metricInc, metricDec},
			FailValues: []tests.Value{metric},
		},
		{
			Predicate:  tests.LT,
			PassValues: []tests.Value{metricDec},
			FailValues: []tests.Value{metric, metricInc},
		},
		{
			Predicate:  tests.LTE,
			PassValues: []tests.Value{metricDec, metric},
			FailValues: []tests.Value{metricInc},
		},
		{
			Predicate:  tests.GT,
			PassValues: []tests.Value{metricInc},
			FailValues: []tests.Value{metric, metricDec},
		},
		{
			Predicate:  tests.GTE,
			PassValues: []tests.Value{metricInc, metric},
			FailValues: []tests.Value{metricDec},
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
	l, r tests.Value,
) {
	t.Helper()
	expectPredicateResult(t, p, l, r, true)
}

func expectPredicateFail(
	t *testing.T,
	p tests.Predicate,
	l, r tests.Value,
) {
	t.Helper()
	expectPredicateResult(t, p, l, r, false)
}

func expectPredicateResult(
	t *testing.T,
	p tests.Predicate,
	l, r tests.Value,
	expPass bool,
) {
	t.Helper()

	if pass := p.Apply(r, l).Pass; pass != expPass {
		t.Errorf("exp %v %d %d -> %v, got %v", p, l, r, expPass, pass)
	}
}
