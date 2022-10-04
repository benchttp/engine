package tests_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/metrics/timestats"
	"github.com/benchttp/engine/runner/internal/tests"
)

func TestRun(t *testing.T) {
	testcases := []struct {
		label          string
		inputAgg       metrics.Aggregate
		inputCases     []tests.Case
		expGlobalPass  bool
		expCaseResults []tests.CaseResult
	}{
		{
			label:    "pass if all cases pass",
			inputAgg: metricsWithMeanResponseTime(ms(100)),
			inputCases: []tests.Case{
				{
					Name:      "average response time below 120ms (pass)",
					Predicate: tests.LT,
					Field:     "ResponseTimes.Mean",
					Target:    ms(120),
				},
				{
					Name:      "average response time is above 80ms (pass)",
					Predicate: tests.GT,
					Field:     "ResponseTimes.Mean",
					Target:    ms(80),
				},
			},
			expGlobalPass: true,
			expCaseResults: []tests.CaseResult{
				{Pass: true, Got: ms(100), Summary: "want ResponseTimes.Mean < 120ms, got 100ms"},
				{Pass: true, Got: ms(100), Summary: "want ResponseTimes.Mean > 80ms, got 100ms"},
			},
		},
		{
			label:    "fail if at least one case fails",
			inputAgg: metricsWithMeanResponseTime(ms(200)),
			inputCases: []tests.Case{
				{
					Name:      "average response time below 120ms (fail)",
					Predicate: tests.LT,
					Field:     "ResponseTimes.Mean",
					Target:    ms(120),
				},
				{
					Name:      "average response time is above 80ms (pass)",
					Predicate: tests.GT,
					Field:     "ResponseTimes.Mean",
					Target:    ms(80),
				},
			},
			expGlobalPass: false,
			expCaseResults: []tests.CaseResult{
				{Pass: false, Got: ms(200), Summary: "want ResponseTimes.Mean < 120ms, got 200ms"},
				{Pass: true, Got: ms(200), Summary: "want ResponseTimes.Mean > 80ms, got 200ms"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			suiteResult := tests.Run(tc.inputAgg, tc.inputCases)

			if gotGlobalPass := suiteResult.Pass; gotGlobalPass != tc.expGlobalPass {
				t.Errorf(
					"exp global pass == %v, got %v",
					gotGlobalPass, tc.expGlobalPass,
				)
			}

			assertEqualCaseResults(t, tc.expCaseResults, suiteResult.Results)
		})
	}
}

func assertEqualCaseResults(t *testing.T, exp, got []tests.CaseResult) {
	t.Helper()

	if gotLen, expLen := len(got), len(exp); gotLen != expLen {
		t.Errorf("exp %d case results, got %d", expLen, gotLen)
	}

	for i, expResult := range exp {
		gotResult := got[i]
		caseDesc := fmt.Sprintf("case #%d (%q)", i, gotResult.Input.Name)

		t.Run(fmt.Sprintf("cases[%d].Pass", i), func(t *testing.T) {
			if gotResult.Pass != expResult.Pass {
				t.Errorf(
					"\n%s:\nexp %v, got %v",
					caseDesc, expResult.Pass, gotResult.Pass,
				)
			}
		})

		t.Run(fmt.Sprintf("cases[%d].Got", i), func(t *testing.T) {
			if gotResult.Got != expResult.Got {
				t.Errorf(
					"\n%s:\nexp %v, got %v",
					caseDesc, expResult.Got, gotResult.Got,
				)
			}
		})

		t.Run(fmt.Sprintf("cases[%d].Summary", i), func(t *testing.T) {
			if gotResult.Summary != expResult.Summary {
				t.Errorf(
					"\n%s:\nexp %q\ngot %q",
					caseDesc, expResult.Summary, gotResult.Summary,
				)
			}
		})
	}
}

func metricsWithMeanResponseTime(d time.Duration) metrics.Aggregate {
	return metrics.Aggregate{
		ResponseTimes: timestats.TimeStats{
			Mean: d,
		},
	}
}

func ms(n int) time.Duration {
	return time.Duration(n) * time.Millisecond
}
