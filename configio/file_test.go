package configio_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/configio"
)

func TestFindFile(t *testing.T) {
	var (
		fileYAML = configPath("valid/benchttp.yaml")
		fileJSON = configPath("valid/benchttp.json")
		nofile   = configPath("does-not-exist.json")
	)

	t.Run("return first existing file from input", func(t *testing.T) {
		files := []string{nofile, fileYAML, fileJSON}

		if got := configio.FindFile(files...); got != fileYAML {
			t.Errorf("did not retrieve good file: exp %s, got %s", fileYAML, got)
		}
	})

	t.Run("return first existing file from defaults", func(t *testing.T) {
		configio.DefaultPaths = []string{nofile, fileYAML, fileJSON}

		if got := configio.FindFile(); got != fileYAML {
			t.Errorf("did not retrieve good file: exp %s, got %s", fileYAML, got)
		}
	})

	t.Run("return empty string when no match", func(t *testing.T) {
		files := []string{nofile}

		if got := configio.FindFile(files...); got != "" {
			t.Errorf("retrieved unexpected file: %s", got)
		}
	})
}

func TestUnmarshalFile(t *testing.T) {
	supportedExt := []string{
		".yml",
		".yaml",
		".json",
	}

	t.Run("return file errors early", func(t *testing.T) {
		testcases := []struct {
			label  string
			path   string
			expErr error
		}{
			{
				label:  "not found",
				path:   configPath("invalid/bad path"),
				expErr: configio.ErrFileNotFound,
			},
			{
				label:  "unsupported extension",
				path:   configPath("invalid/badext.yams"),
				expErr: configio.ErrFileExt,
			},
			{
				label:  "yaml invalid fields",
				path:   configPath("invalid/badfields.yml"),
				expErr: configio.ErrFileParse,
			},
			{
				label:  "json invalid fields",
				path:   configPath("invalid/badfields.json"),
				expErr: configio.ErrFileParse,
			},
			{
				label:  "self reference",
				path:   configPath("extends/extends-circular-self.yml"),
				expErr: configio.ErrFileCircular,
			},
			{
				label:  "circular reference",
				path:   configPath("extends/extends-circular-0.yml"),
				expErr: configio.ErrFileCircular,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				runner := benchttp.Runner{}
				gotErr := configio.UnmarshalFile(tc.path, &runner)

				if gotErr == nil {
					t.Fatal("exp non-nil error, got nil")
				}

				if !errors.Is(gotErr, tc.expErr) {
					t.Errorf("\nexp %v\ngot %v", tc.expErr, gotErr)
				}

				if !sameRunner(runner, benchttp.Runner{}) {
					t.Errorf("\nexp empty config\ngot %v", runner)
				}
			})
		}
	})

	t.Run("happy path for all extensions", func(t *testing.T) {
		for _, ext := range supportedExt {
			filename := configPath("valid/benchttp" + ext)
			runner := benchttp.Runner{}

			if err := configio.UnmarshalFile(filename, &runner); err != nil {
				// critical error, stop the test
				t.Fatal(err)
			}

			if sameRunner(runner, benchttp.Runner{}) {
				t.Error("received an empty configuration")
			}

			exp := expectedRunner()
			if !sameRunner(runner, exp) {
				t.Errorf("unexpected parsed config for %s file:\nexp %#v\ngot %#v", ext, exp, runner)
			}

		}
	})

	t.Run("override input config", func(t *testing.T) {
		filename := configPath("valid/benchttp-zeros.yml")
		runner := benchttp.Runner{}
		runner.Request = httptest.NewRequest("POST", "https://overriden.com", nil)
		runner.GlobalTimeout = 10 * time.Millisecond

		if err := configio.UnmarshalFile(filename, &runner); err != nil {
			t.Fatal(err)
		}

		const (
			expMethod        = "POST"                // from input config
			expGlobalTimeout = 42 * time.Millisecond // from read file
		)

		if gotMethod := runner.Request.Method; gotMethod != expMethod {
			t.Errorf(
				"did not keep input values that are not set: "+
					"exp Request.Method == %s, got %s",
				expMethod, gotMethod,
			)
		}

		if gotGlobalTimeout := runner.GlobalTimeout; gotGlobalTimeout != expGlobalTimeout {
			t.Errorf(
				"did not override input values that are set: "+
					"exp Runner.GlobalTimeout == %v, got %v",
				expGlobalTimeout, gotGlobalTimeout,
			)
		}

		t.Log(runner)
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
				var runner benchttp.Runner
				if err := configio.UnmarshalFile(tc.cfpath, &runner); err != nil {
					t.Fatal(err)
				}

				var (
					expMethod = "POST"
					expURL    = fmt.Sprintf("http://%s.config", tc.cfname)
				)

				if gotMethod := runner.Request.Method; gotMethod != expMethod {
					t.Errorf("method: exp %s, got %s", expMethod, gotMethod)
				}

				if gotURL := runner.Request.URL.String(); gotURL != expURL {
					t.Errorf("url: exp %s, got %s", expURL, gotURL)
				}
			})
		}
	})
}

// helpers

// expectedRunner returns the expected benchttp.Runner after unmarhsaling
// one of the valid config files in testdata.
func expectedRunner() benchttp.Runner {
	request := httptest.NewRequest(
		"POST",
		"http://localhost:9999?fib=30&delay=200ms",
		bytes.NewReader([]byte(`{"key0":"val0","key1":"val1"}`)),
	)
	request.Header = http.Header{
		"key0": []string{"val0", "val1"},
		"key1": []string{"val0"},
	}
	return benchttp.Runner{
		Request: request,

		Requests:       100,
		Concurrency:    1,
		Interval:       50 * time.Millisecond,
		RequestTimeout: 2 * time.Second,
		GlobalTimeout:  60 * time.Second,

		Tests: []benchttp.TestCase{
			{
				Name:      "minimum response time",
				Field:     "ResponseTimes.Min",
				Predicate: "GT",
				Target:    80 * time.Millisecond,
			},
			{
				Name:      "maximum response time",
				Field:     "ResponseTimes.Max",
				Predicate: "LTE",
				Target:    120 * time.Millisecond,
			},
			{
				Name:      "100% availability",
				Field:     "RequestFailureCount",
				Predicate: "EQ",
				Target:    0,
			},
		},
	}
}

func sameRunner(a, b benchttp.Runner) bool {
	if a.Request == nil || b.Request == nil {
		return a.Request == nil && b.Request == nil
	}
	return sameURL(a.Request.URL, b.Request.URL) &&
		reflect.DeepEqual(a.Request.Header, b.Request.Header) &&
		reflect.DeepEqual(a.Request.Body, b.Request.Body)
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
		restore := setTempValue(&u.RawQuery, "replaced by test")
		defer restore()
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
	return filepath.Join("testdata", name)
}
