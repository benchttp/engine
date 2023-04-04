package render_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/benchttp/engine/benchttp"
	"github.com/benchttp/engine/benchttp/metrics"

	"github.com/benchttp/engine/cli/render"
	"github.com/benchttp/engine/cli/render/ansi"
)

func TestReport_String(t *testing.T) {
	t.Run("returns metrics summary", func(t *testing.T) {
		metrics, duration := metricsStub()
		runner := runnerStub()

		rep := &benchttp.Report{
			Metrics: metrics,
			Metadata: benchttp.Metadata{
				Runner:        runner,
				TotalDuration: duration,
			},
		}
		checkSummary(t, render.ReportSummaryString(rep))
	})
}

// helpers

func metricsStub() (agg metrics.Aggregate, total time.Duration) {
	return metrics.Aggregate{
		RequestFailures: make([]struct {
			Reason string
		}, 1),
		Records: make([]struct{ ResponseTime time.Duration }, 3),
		ResponseTimes: metrics.TimeStats{
			Min:  4 * time.Second,
			Max:  6 * time.Second,
			Mean: 5 * time.Second,
		},
	}, 15 * time.Second
}

func runnerStub() benchttp.Runner {
	runner := benchttp.Runner{}
	runner.Request = mustMakeRequest("https://a.b.com")
	runner.Requests = -1
	return runner
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

func mustMakeRequest(uri string) *http.Request {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	return req
}
