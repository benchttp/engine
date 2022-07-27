package output_test

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/output"
)

func TestReport_String(t *testing.T) {
	const d = 15 * time.Second

	t.Run("return default summary if template is empty", func(t *testing.T) {
		const tpl = ""

		rep := output.New(newMetrics(), newConfigWithTemplate(tpl), d)
		checkSummary(t, rep.String())
	})

	t.Run("return executed template if valid", func(t *testing.T) {
		const tpl = "{{ .Metrics.TotalCount }}"

		m := newMetrics()
		rep := output.New(m, newConfigWithTemplate(tpl), d)

		if got, exp := rep.String(), strconv.Itoa(m.TotalCount); got != exp {
			t.Errorf("\nunexpected output\nexp %s\ngot %s", exp, got)
		}
	})

	t.Run("fallback to default summary if template is invalid", func(t *testing.T) {
		const tpl = "{{ .Marcel.Patulacci }}"

		rep := output.New(newMetrics(), newConfigWithTemplate(tpl), d)
		got := rep.String()
		split := strings.Split(got, "Falling back to default summary:\n")

		if len(split) != 2 {
			t.Fatalf("\nunexpected output:\n%s", got)
		}

		errMsg, summary := split[0], split[1]
		if !strings.Contains(errMsg, "template syntax error") {
			t.Errorf("\nexp template syntax error\ngot %s", errMsg)
		}

		checkSummary(t, summary)
	})
}

// helpers

func newMetrics() metrics.Aggregate {
	return metrics.Aggregate{
		FailureCount: 1,
		SuccessCount: 2,
		TotalCount:   3,
		Min:          4 * time.Second,
		Max:          6 * time.Second,
		Avg:          5 * time.Second,
	}
}

func newConfigWithTemplate(tpl string) config.Global {
	urlURL, _ := url.ParseRequestURI("https://a.b.com")
	return config.Global{
		Request: config.Request{URL: urlURL},
		Runner:  config.Runner{Requests: -1},
		Output:  config.Output{Template: tpl},
	}
}

func checkSummary(t *testing.T, summary string) {
	t.Helper()

	expSummary := `
Endpoint           https://a.b.com
Requests           3/∞
Errors             1
Min response time  4000ms
Max response time  6000ms
Mean response time 5000ms
Total duration     15000ms
`[1:]

	if summary != expSummary {
		t.Errorf("\nexp summary:\n%q\ngot summary:\n%q", expSummary, summary)
	}
}
