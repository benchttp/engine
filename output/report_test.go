package output_test

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

func TestReport_String(t *testing.T) {
	t.Run("return default summary if template is empty", func(t *testing.T) {
		const tpl = ""

		rep := output.New(newBenchmark(), newConfigWithOutput(tpl))
		checkSummary(t, rep.String())
	})

	t.Run("return executed template if valid", func(t *testing.T) {
		const tpl = "{{ .Benchmark.Length }}"

		bk := newBenchmark()
		rep := output.New(bk, newConfigWithOutput(tpl))

		if got, exp := rep.String(), strconv.Itoa(bk.Length); got != exp {
			t.Errorf("\nunexpected output\nexp %s\ngot %s", exp, got)
		}
	})

	t.Run("fallback to default summary if template is invalid", func(t *testing.T) {
		const tpl = "{{ .Marcel.Patulacci }}"

		rep := output.New(newBenchmark(), newConfigWithOutput(tpl))
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

func TestReport_HTTPRequest(t *testing.T) {
	t.Run("generate gob-encoded POST request to target endpoint", func(t *testing.T) {
		bk, cfg := newBenchmark(), newConfigWithOutput("")
		rep := output.New(bk, cfg)

		req, err := rep.HTTPRequest()
		if err != nil {
			t.Fatalf("unexpected error:\n%s", err)
		}
		checkRequest(t, req, rep)
	})
}

func TestReport_Export(t *testing.T) {
	t.Run("return ErrInvalidStrategy if Strategy is invalid", func(t *testing.T) {
		rep := output.New(newBenchmark(), newConfigWithOutput("", "nostrat"))
		if gotErr := rep.Export(); !errors.Is(gotErr, output.ErrInvalidStrategy) {
			t.Errorf("\nexp ErrInvalidStrategy\ngot %v", gotErr)
		}
	})

	// TODO:
	// - "return accumulated errors"
	// - "happy path"
	// Both tests require to mock package export which implies
	// increased complexity to its implementation.
	// Let's keep that for later.
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

func newConfigWithOutput(tpl string, strats ...config.OutputStrategy) config.Global {
	urlURL, _ := url.ParseRequestURI("https://a.b.com")
	return config.Global{
		Request: config.Request{URL: urlURL},
		Runner:  config.Runner{Requests: -1},
		Output: config.Output{
			Out:      strats,
			Template: tpl,
		},
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

func checkRequest(t *testing.T, r *http.Request, rep *output.Report) {
	t.Helper()

	if r == nil {
		t.Fatal("returned nil request")
	}

	t.Run("set method to POST", func(t *testing.T) {
		const expMethod = "POST"
		if r.Method != expMethod {
			t.Errorf("request method: exp %q, got %q", expMethod, r.Method)
		}
	})

	t.Run("set target url to correct endpoint", func(t *testing.T) {
		const expURL = "http://localhost:9998/v1/report"
		if gotURL := fmt.Sprint(r.URL); gotURL != expURL {
			t.Errorf("request method: exp %q, got %q", expURL, gotURL)
		}
	})

	t.Run("set body to gob encoded report", func(t *testing.T) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		gotReport := &output.Report{}
		if err := gob.NewDecoder(bytes.NewReader(b)).Decode(gotReport); err != nil {
			t.Fatal(err)
		}
		if !sameReports(rep, gotReport) {
			t.Errorf("unexpected report:\nexp %#v\ngot %#v", rep, gotReport)
		}
	})
}

// sameReports performs a deep equality check for two output.Report,
// ignoring Report.log value.
func sameReports(a, b *output.Report) bool {
	// cannot rely on reflect.DeepEqual for time.Time
	// https://github.com/golang/go/issues/10089
	if !a.Metadata.FinishedAt.Equal(b.Metadata.FinishedAt) {
		return false
	}

	// times are equal, use exact same value for both reports
	// so we can perform deep equality check
	bMetadata := b.Metadata
	bMetadata.FinishedAt = a.Metadata.FinishedAt

	// exclude Report.log func
	return reflect.DeepEqual(
		output.Report{
			Benchmark: a.Benchmark,
			Metadata:  a.Metadata,
		},
		output.Report{
			Benchmark: b.Benchmark,
			Metadata:  bMetadata,
		},
	)
}
