package configparse_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

const (
	testdataConfigPath = "./testdata"
	testURL            = "http://localhost:9999?fib=30&delay=200ms"
)

var supportedExt = []string{
	".yml",
	".yaml",
	".json",
}

// TestParse ensures the config file is open, read, and correctly parsed.
func TestParse(t *testing.T) {
	t.Run("return file errors early", func(t *testing.T) {
		testcases := []struct {
			label  string
			path   string
			expErr error
		}{
			{
				label:  "not found",
				path:   configPath("invalid/bad path"),
				expErr: configparse.ErrFileNotFound,
			},
			{
				label:  "unsupported extension",
				path:   configPath("invalid/badext.yams"),
				expErr: configparse.ErrFileExt,
			},
			{
				label:  "yaml invalid fields",
				path:   configPath("invalid/badfields.yml"),
				expErr: configparse.ErrParse,
			},
			{
				label:  "json invalid fields",
				path:   configPath("invalid/badfields.json"),
				expErr: configparse.ErrParse,
			},
			{
				label:  "self reference",
				path:   configPath("extends/extends-circular-self.yml"),
				expErr: configparse.ErrCircularExtends,
			},
			{
				label:  "circular reference",
				path:   configPath("extends/extends-circular-0.yml"),
				expErr: configparse.ErrCircularExtends,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				gotCfg, gotErr := configparse.Parse(tc.path)

				if gotErr == nil {
					t.Fatal("exp non-nil error, got nil")
				}

				if !errors.Is(gotErr, tc.expErr) {
					t.Errorf("\nexp %v\ngot %v", tc.expErr, gotErr)
				}

				if !reflect.DeepEqual(gotCfg, runner.Config{}) {
					t.Errorf("\nexp empty config\ngot %v", gotCfg)
				}
			})
		}
	})

	t.Run("happy path for all extensions", func(t *testing.T) {
		for _, ext := range supportedExt {
			expCfg := newExpConfig()
			fname := configPath("valid/benchttp" + ext)

			gotCfg, err := configparse.Parse(fname)
			if err != nil {
				// critical error, stop the test
				t.Fatal(err)
			}

			expURL, gotURL := expCfg.Request.URL, gotCfg.Request.URL

			// compare *url.URLs separately, as they contain unpredictable values
			// they need special checks
			if !sameURL(gotURL, expURL) {
				t.Errorf("unexpected parsed URL:\nexp %v, got %v", expURL, gotURL)
			}

			// replace unpredictable values (undetermined query params order)
			restoreGotCfg := setTempValue(&gotURL.RawQuery, "replaced by test")
			restoreExpCfg := setTempValue(&expURL.RawQuery, "replaced by test")

			if !reflect.DeepEqual(gotCfg, expCfg) {
				t.Errorf("unexpected parsed config for %s file:\nexp %v\ngot %v", ext, expCfg, gotCfg)
			}

			restoreExpCfg()
			restoreGotCfg()
		}
	})

	t.Run("override default values", func(t *testing.T) {
		const (
			expRequests      = 0 // default is -1
			expGlobalTimeout = 42 * time.Millisecond
		)

		fname := configPath("valid/benchttp-zeros.yml")

		cfg, err := configparse.Parse(fname)
		if err != nil {
			t.Fatal(err)
		}

		if gotRequests := cfg.Runner.Requests; gotRequests != expRequests {
			t.Errorf("did not override Requests: exp %d, got %d", expRequests, gotRequests)
		}

		if gotGlobalTimeout := cfg.Runner.GlobalTimeout; gotGlobalTimeout != expGlobalTimeout {
			t.Errorf("did not override GlobalTimeout: exp %d, got %d", expGlobalTimeout, gotGlobalTimeout)
		}

		t.Log(cfg)
	})

	t.Run("extend config files", func(t *testing.T) {
		testcases := []struct {
			label  string
			cfname string
			cfpath string
		}{
			{
				label:  "same directory",
				cfname: "child",
				cfpath: configPath("extends/extends-valid-child.yml"),
			},
			{
				label:  "nested directory",
				cfname: "nested",
				cfpath: configPath("extends/nest-0/nest-1/extends-valid-nested.yml"),
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				cfg, err := configparse.Parse(tc.cfpath)
				if err != nil {
					t.Fatal(err)
				}

				var (
					expMethod = "POST"
					expURL    = fmt.Sprintf("http://%s.config", tc.cfname)
				)

				if gotMethod := cfg.Request.Method; gotMethod != expMethod {
					t.Errorf("method: exp %s, got %s", expMethod, gotMethod)
				}

				if gotURL := cfg.Request.URL.String(); gotURL != expURL {
					t.Errorf("method: exp %s, got %s", expURL, gotURL)
				}
			})
		}
	})
}

// helpers

// newExpConfig returns the expected runner.ConfigConfig result after parsing
// one of the config files in testdataConfigPath.
func newExpConfig() runner.Config {
	u, _ := url.ParseRequestURI(testURL)
	return runner.Config{
		Request: runner.RequestConfig{
			Method: "POST",
			URL:    u,
			Header: http.Header{
				"key0": []string{"val0", "val1"},
				"key1": []string{"val0"},
			},
			Body: runner.NewRequestBody("raw", `{"key0":"val0","key1":"val1"}`),
		},
		Runner: runner.RecorderConfig{
			Requests:       100,
			Concurrency:    1,
			Interval:       50 * time.Millisecond,
			RequestTimeout: 2 * time.Second,
			GlobalTimeout:  60 * time.Second,
		},
		Output: runner.OutputConfig{
			Silent:   true,
			Template: "{{ .Metrics.Avg }}",
		},
		Tests: []runner.TestCase{
			{
				Name:      "minimum response time",
				Source:    "MIN",
				Predicate: "GT",
				Target:    80 * time.Millisecond,
			},
			{
				Name:      "maximum response time",
				Source:    "MAX",
				Predicate: "LTE",
				Target:    120 * time.Millisecond,
			},
		},
	}
}

// sameURL returns true if a and b are the same *url.URL, taking into account
// the undeterministic nature of their RawQuery.
func sameURL(a, b *url.URL) bool {
	// check query params equality via Query() rather than RawQuery
	if !reflect.DeepEqual(a.Query(), b.Query()) {
		return false
	}

	// temporarily set RawQuery to a determined value
	for _, u := range []*url.URL{a, b} {
		defer setTempValue(&u.RawQuery, "replaced by test")()
	}

	// we can now rely on deep equality check
	return reflect.DeepEqual(a, b)
}

// setTempValue sets *ptr to val and returns a restore func that sets *ptr
// back to its previous value.
func setTempValue(ptr *string, val string) (restore func()) {
	previousValue := *ptr
	*ptr = val
	return func() {
		*ptr = previousValue
	}
}

func configPath(name string) string {
	return filepath.Join(testdataConfigPath, name)
}
