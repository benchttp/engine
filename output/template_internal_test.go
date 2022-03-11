package output

import (
	"errors"
	"testing"
	"time"

	"github.com/benchttp/runner/requester"
)

func TestReport_applyTemplate(t *testing.T) {
	testcases := []struct {
		label   string
		pattern string
		expStr  string
		expErr  error
	}{
		{
			label:   "return errTemplateEmpty if pattern is empty",
			pattern: "",
			expStr:  "",
			expErr:  errTemplateEmpty,
		},
		{
			label:   "return errTemplateSyntaxt if pattern has syntax error",
			pattern: "{{ else }}",
			expStr:  "",
			expErr:  errTemplateSyntax,
		},
		{
			label:   "return errTemplateSyntaxt if pattern doesn't match report values",
			pattern: "{{ .Foo }}", // Report.Foo doesn't exist
			expStr:  "",
			expErr:  errTemplateSyntax,
		},
		{
			label:   "happy path with custom template functions",
			pattern: "{{ stats.Min }},{{ stats.Max }},{{ stats.Mean }}",
			expStr:  "0s,0s,0s",
			expErr:  nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			r := &Report{}
			gotStr, gotErr := r.applyTemplate(tc.pattern)
			if !errors.Is(gotErr, tc.expErr) {
				t.Errorf("unexpected error: %v", gotErr)
			}
			if gotStr != tc.expStr {
				t.Errorf("unexpected string: %q", gotStr)
			}
		})
	}
}

func TestReport_templateFuncs(t *testing.T) {
	t.Run("stats", func(t *testing.T) {
		rep := newFilledReport()

		v := retrieveTemplateFuncOrFatal(t, rep, "stats")

		f, ok := v.(func() basicStats)
		if !ok {
			t.Fatalf("wrong type:\nexp func() basicStats\ngot %T", v)
		}

		if gotStats := f(); (gotStats != basicStats{
			Min:  1 * time.Second,
			Max:  3 * time.Second,
			Mean: 2 * time.Second,
		}) {
			t.Errorf("unexpected stats: %+v", gotStats)
		}
	})

	t.Run("event", func(t *testing.T) {
		rep := newFilledReport()

		v := retrieveTemplateFuncOrFatal(t, rep, "event")

		f, ok := v.(func(requester.Record, string) time.Duration)
		if !ok {
			t.Fatalf("wrong type:\nexp func(requester.Record, string) time.Duration\ngot %T", v)
		}

		t.Run("return matching event", func(t *testing.T) {
			rec := rep.Benchmark.Records[0]
			if got, exp := f(rec, "event1"), rec.Events[1].Time; got != exp {
				t.Errorf("unexpected time: exp %s, got %s", exp, got)
			}
		})

		t.Run("return 0 if no match", func(t *testing.T) {
			rec := rep.Benchmark.Records[0]
			if got, exp := f(rec, "nomatch"), time.Duration(0); got != exp {
				t.Errorf("unexpected time: exp %s, got %s", exp, got)
			}
		})
	})

	t.Run("fail", func(t *testing.T) {
		rep := newFilledReport()

		v := retrieveTemplateFuncOrFatal(t, rep, "fail")

		f, ok := v.(func(...interface{}) string)
		if !ok {
			t.Fatalf("wrong type:\nexp func(...interface{}) string\ngot %T", v)
		}

		if got := f("a", "b", "c"); got != "" {
			t.Errorf("unexpected output: exp always %q, got %q", "", got)
		}

		gotErr := rep.errTemplateFailTriggered
		if !errors.Is(gotErr, ErrTemplateFailTriggered) {
			t.Fatalf("unexpected error:\nexp ErrTemplateFailTriggered\ngot %v", gotErr)
		}
		if gotMsg, expMsg := gotErr.Error(), "test failed: abc"; gotMsg != expMsg {
			t.Errorf("unexpected error message:\nexp %q\ngot %q", expMsg, gotMsg)
		}
	})
}

// helpers

func retrieveTemplateFuncOrFatal(t *testing.T, r *Report, name string) interface{} {
	t.Helper()
	v, exists := r.templateFuncs()[name]
	if !exists {
		t.Fatalf("template func %q does not exist", name)
	}
	return v
}

// newFilledReport returns a new report with some values set.
func newFilledReport() *Report {
	return &Report{
		Benchmark: requester.Benchmark{
			Records: []requester.Record{
				{
					Time: 1 * time.Second,
					Events: []requester.Event{
						{Name: "event0", Time: 400 * time.Millisecond},
						{Name: "event1", Time: 600 * time.Millisecond},
					},
				},
				{
					Time: 3 * time.Second,
					Events: []requester.Event{
						{Name: "event0", Time: 2 * time.Second},
						{Name: "event1", Time: 1 * time.Second},
					},
				},
			},
		},
	}
}
