package output_test

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/engine/config"
	"github.com/benchttp/engine/output"
	"github.com/benchttp/engine/requester"
)

func TestReport_String(t *testing.T) {
	t.Run("return default summary if template is empty", func(t *testing.T) {
		const tpl = ""

		rep := output.New(newBenchmark(), newConfigWithTemplate(tpl))
		checkSummary(t, rep.String())
	})

	t.Run("return executed template if valid", func(t *testing.T) {
		const tpl = "{{ .Benchmark.Length }}"

		bk := newBenchmark()
		rep := output.New(bk, newConfigWithTemplate(tpl))

		if got, exp := rep.String(), strconv.Itoa(bk.Length); got != exp {
			t.Errorf("\nunexpected output\nexp %s\ngot %s", exp, got)
		}
	})

	t.Run("fallback to default summary if template is invalid", func(t *testing.T) {
		const tpl = "{{ .Marcel.Patulacci }}"

		rep := output.New(newBenchmark(), newConfigWithTemplate(tpl))
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

func newBenchmark() requester.Benchmark {
	return requester.Benchmark{
		Fail:     1,
		Success:  2,
		Length:   3,
		Duration: 4 * time.Second,
		Records: []requester.Record{
			{Time: 5 * time.Second},
			{Time: 6 * time.Second},
			{Time: 7 * time.Second},
		},
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
Min response time  5000ms
Max response time  7000ms
Mean response time 6000ms
Total duration     4000ms
`[1:]

	if summary != expSummary {
		t.Errorf("\nexp summary:\n%q\ngot summary:\n%q", expSummary, summary)
	}
}
