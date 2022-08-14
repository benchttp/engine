package report_test

import (
	"testing"
	"time"

	"github.com/benchttp/engine/internal/cli/ansi"
	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/report"
	"github.com/benchttp/engine/runner/internal/tests"
	"github.com/benchttp/engine/runner/internal/timestats"
)

func TestReport_String(t *testing.T) {
	t.Run("returns metrics summary", func(t *testing.T) {
		agg, d := metricsStub()
		cfg := configStub()

		rep := report.New(agg, cfg, d, tests.SuiteResult{})
		checkSummary(t, rep.String())
	})
}

// helpers

func metricsStub() (agg metrics.Aggregate, total time.Duration) {
	return metrics.Aggregate{
		RequestFailures: make([]struct {
			Reason string
		}, 1),
		Records: make([]struct{ ResponseTime time.Duration }, 3),
		ResponseTimes: timestats.TimeStats{
			Min: 4 * time.Second,
			Max: 6 * time.Second,
			Avg: 5 * time.Second,
		},
	}, 15 * time.Second
}

func configStub() config.Global {
	cfg := config.Global{}
	cfg.Request = cfg.Request.WithURL("https://a.b.com")
	cfg.Runner.Requests = -1
	return cfg
}

func checkSummary(t *testing.T, summary string) {
	t.Helper()

	expSummary := ansi.Bold("→ Summary") + `
Endpoint           https://a.b.com
Requests           3/∞
Errors             1
Min response time  4000ms
Max response time  6000ms
Mean response time 5000ms
Total duration     15000ms
`

	if summary != expSummary {
		t.Errorf("\nexp summary:\n%q\ngot summary:\n%q", expSummary, summary)
	}
}
