package output

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/output/export"
	"github.com/benchttp/runner/requester"
)

func TestReport_Export(t *testing.T) {
	t.Cleanup(mockExportFuncs(false, false))

	testcases := []struct {
		label      string
		strategies []config.OutputStrategy
		userToken  string
		failJSON   bool
		failHTTP   bool
		expErr     string
	}{
		{
			label:      "return ErrInvalidStrategy if Strategy is invalid",
			strategies: []config.OutputStrategy{"nostrat"},
			userToken:  "",
			failJSON:   true,
			failHTTP:   true,
			expErr:     ErrInvalidStrategy.Error(),
		},
		{
			label:      "return JSON error if exportJSONFile fails",
			strategies: []config.OutputStrategy{config.OutputJSON},
			userToken:  "",
			failJSON:   true,
			failHTTP:   true,
			expErr:     "output:\n  - JSON error",
		},
		{
			label:      "return HTTP error if exportHTTP fails",
			strategies: []config.OutputStrategy{config.OutputBenchttp},
			userToken:  "abc",
			failJSON:   true,
			failHTTP:   true,
			expErr:     "output:\n  - HTTP error",
		},
		{
			label:      "return ErrNoToken if exportHTTP is used without a token",
			strategies: []config.OutputStrategy{config.OutputBenchttp},
			userToken:  "",
			failJSON:   true,
			failHTTP:   true,
			expErr:     "output:\n  - user token not set",
		},
		{
			label: "return cumulated errors",
			strategies: []config.OutputStrategy{
				config.OutputJSON,
				config.OutputBenchttp,
			},
			userToken: "abc",
			failJSON:  true,
			failHTTP:  true,
			expErr:    "output:\n  - JSON error\n  - HTTP error",
		},
		{
			label: "happy path",
			strategies: []config.OutputStrategy{
				config.OutputStdout,
				config.OutputJSON,
				config.OutputBenchttp,
			},
			userToken: "abc",
			failJSON:  false,
			failHTTP:  false,
			expErr:    "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			mockExportFuncs(tc.failJSON, tc.failHTTP)

			cfg := newConfigWithStrat(tc.strategies...)
			rep := New(requester.Benchmark{}, cfg, tc.userToken)

			if err := rep.Export(); err != nil && err.Error() != tc.expErr {
				t.Errorf("unexpected error:\nexp %q\ngot %q", tc.expErr, err)
			}
		})
	}

	t.Run("return triggered template error", func(t *testing.T) {
		mockExportFuncs(true, true)

		cfg := newConfigWithStrat(config.OutputStdout)
		rep := New(requester.Benchmark{}, cfg, "")
		rep.errTemplateFailTriggered = ErrTemplateFailTriggered

		if err := rep.Export(); !errors.Is(err, ErrTemplateFailTriggered) {
			t.Errorf(
				"unexpected error:\nexp ErrTemplateFailTriggered\ngot %v", err,
			)
		}
	})
}

func TestGenFilename(t *testing.T) {
	testcases := []struct {
		label string
		in    time.Time
		exp   string
	}{
		{
			label: "return timestamped filename",
			in:    time.Date(1234, time.December, 13, 14, 15, 16, 17, time.UTC),
			exp:   "./benchttp.report.12341213141516.json",
		},
		{
			label: "return timestamped filename with added zeros",
			in:    time.Date(1, time.January, 1, 1, 1, 1, 1, time.UTC),
			exp:   "./benchttp.report.00010101010101.json",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			got := genFilename(tc.in)
			if got != tc.exp {
				t.Errorf("\nexp %s\ngot %s", tc.exp, got)
			}
		})
	}
}

// helpers

var (
	origExportStdout   = exportStdout
	origExportJSONFile = exportJSONFile
	origExportHTTP     = exportHTTP
)

// mockExportFuncs mocks the functions from package export
// to avoid side effects and return a function to restore
// their initial value.
func mockExportFuncs(failJSON, failHTTP bool) (restore func()) {
	exportStdout = func(fmt.Stringer) {}

	exportJSONFile = func(string, interface{}) error {
		if failJSON {
			return errors.New("JSON error")
		}
		return nil
	}

	exportHTTP = func(export.HTTPRequester) error {
		if failHTTP {
			return errors.New("HTTP error")
		}
		return nil
	}

	return func() {
		exportStdout = origExportStdout
		exportJSONFile = origExportJSONFile
		exportHTTP = origExportHTTP
	}
}

// newConfigWithStrat returns a config.Global initialized with the given
// output strategies.
func newConfigWithStrat(strats ...config.OutputStrategy) config.Global {
	urlURL, _ := url.ParseRequestURI("https://a.b.com")
	return config.Global{
		Request: config.Request{URL: urlURL},
		Runner:  config.Runner{Requests: -1},
		Output: config.Output{
			Out:    strats,
			Silent: true,
		},
	}
}
