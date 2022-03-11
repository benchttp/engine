package config_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
)

var validBody = config.NewBody("raw", `{"key0": "val0", "key1": "val1"}`)

func TestGlobal_Validate(t *testing.T) {
	t.Run("return nil if config is valid", func(t *testing.T) {
		cfg := config.Global{
			Request: config.Request{
				Body: validBody,
			}.WithURL("https://github.com/benchttp/"),
			Runner: config.Runner{
				Requests:       5,
				Concurrency:    5,
				Interval:       5,
				RequestTimeout: 5,
				GlobalTimeout:  5,
			},
			Output: config.Output{
				Out: []config.OutputStrategy{"stdout", "json", "benchttp"},
			},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("return cumulated errors if config is invalid", func(t *testing.T) {
		cfg := config.Global{
			Request: config.Request{
				Body: config.Body{},
			}.WithURL("abc"),
			Runner: config.Runner{
				Requests:       -5,
				Concurrency:    -5,
				Interval:       -5,
				RequestTimeout: -5,
				GlobalTimeout:  -5,
			},
			Output: config.Output{
				Out: []config.OutputStrategy{config.OutputStdout, "bad-output"},
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Fatal("invalid configuration considered valid")
		}

		var errInvalid *config.InvalidConfigError
		if !errors.As(err, &errInvalid) {
			t.Fatalf("unexpected error type: %T", err)
		}

		errs := errInvalid.Errors
		findErrorOrFail(t, errs, `url (""): invalid`)
		findErrorOrFail(t, errs, `requests (-5): want >= 0`)
		findErrorOrFail(t, errs, `concurrency (-5): want > 0 and <= requests (-5)`)
		findErrorOrFail(t, errs, `interval (-5): want >= 0`)
		findErrorOrFail(t, errs, `requestTimeout (-5): want > 0`)
		findErrorOrFail(t, errs, `globalTimeout (-5): want > 0`)
		findErrorOrFail(t, errs, `out ("bad-output"): want one or many of "benchttp", "json", "stdout"`)

		t.Logf("got error:\n%v", errInvalid)
	})
}

func TestGlobal_Override(t *testing.T) {
	t.Run("do not override unspecified fields", func(t *testing.T) {
		baseCfg := config.Global{}
		newCfg := config.Global{
			Request: config.Request{
				Body: config.Body{},
			}.WithURL("http://a.b?p=2"),
			Runner: config.Runner{
				Requests:       1,
				Concurrency:    2,
				RequestTimeout: 3 * time.Second,
				GlobalTimeout:  4 * time.Second,
			},
			Output: config.Output{
				Out:    []config.OutputStrategy{config.OutputStdout},
				Silent: true,
			},
		}

		if gotCfg := baseCfg.Override(newCfg); !reflect.DeepEqual(gotCfg, baseCfg) {
			t.Errorf("overrode unexpected fields:\nexp %#v\ngot %#v", baseCfg, gotCfg)
		}
	})

	t.Run("override specified fields", func(t *testing.T) {
		baseCfg := config.Global{}
		newCfg := config.Global{
			Request: config.Request{
				Body: validBody,
			}.WithURL("http://a.b?p=2"),
			Runner: config.Runner{
				Requests:       1,
				Concurrency:    2,
				RequestTimeout: 3 * time.Second,
				GlobalTimeout:  4 * time.Second,
			},
			Output: config.Output{
				Out:    []config.OutputStrategy{config.OutputStdout},
				Silent: true,
			},
		}
		fields := []string{
			config.FieldMethod,
			config.FieldURL,
			config.FieldRequests,
			config.FieldConcurrency,
			config.FieldRequestTimeout,
			config.FieldGlobalTimeout,
			config.FieldBody,
			config.FieldOut,
			config.FieldSilent,
		}

		if gotCfg := baseCfg.Override(newCfg, fields...); !reflect.DeepEqual(gotCfg, newCfg) {
			t.Errorf("did not override expected fields:\nexp %v\ngot %v", baseCfg, gotCfg)
			t.Log(fields)
		}
	})

	t.Run("override header selectively", func(t *testing.T) {
		testcases := []struct {
			label     string
			oldHeader http.Header
			newHeader http.Header
			expHeader http.Header
		}{
			{
				label:     "erase overridden keys",
				oldHeader: http.Header{"key": []string{"oldval"}},
				newHeader: http.Header{"key": []string{"newval"}},
				expHeader: http.Header{"key": []string{"newval"}},
			},
			{
				label:     "do not erase not overridden keys",
				oldHeader: http.Header{"key": []string{"oldval"}},
				newHeader: http.Header{},
				expHeader: http.Header{"key": []string{"oldval"}},
			},
			{
				label:     "add new keys",
				oldHeader: http.Header{"key0": []string{"oldval"}},
				newHeader: http.Header{"key1": []string{"newval"}},
				expHeader: http.Header{
					"key0": []string{"oldval"},
					"key1": []string{"newval"},
				},
			},
			{
				label: "erase only overridden keys",
				oldHeader: http.Header{
					"key0": []string{"oldval0", "oldval1"},
					"key1": []string{"oldval0", "oldval1"},
				},
				newHeader: http.Header{
					"key1": []string{"newval0", "newval1"},
					"key2": []string{"newval0", "newval1"},
				},
				expHeader: http.Header{
					"key0": []string{"oldval0", "oldval1"},
					"key1": []string{"newval0", "newval1"},
					"key2": []string{"newval0", "newval1"},
				},
			},
			{
				label:     "nil new header does nothing",
				oldHeader: http.Header{"key": []string{"val"}},
				newHeader: nil,
				expHeader: http.Header{"key": []string{"val"}},
			},
			{
				label:     "replace nil old header",
				oldHeader: nil,
				newHeader: http.Header{"key": []string{"val"}},
				expHeader: http.Header{"key": []string{"val"}},
			},
			{
				label:     "nil over nil is nil",
				oldHeader: nil,
				newHeader: nil,
				expHeader: nil,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				oldCfg := config.Global{
					Request: config.Request{
						Header: tc.oldHeader,
					},
				}

				newCfg := config.Global{
					Request: config.Request{
						Header: tc.newHeader,
					},
				}

				gotCfg := oldCfg.Override(newCfg, config.FieldHeader)

				if gotHeader := gotCfg.Request.Header; !reflect.DeepEqual(gotHeader, tc.expHeader) {
					t.Errorf("\nexp %#v\ngot %#v", tc.expHeader, gotHeader)
				}
			})
		}
	})
}

func TestRequest_WithURL(t *testing.T) {
	t.Run("set empty url if invalid", func(t *testing.T) {
		cfg := config.Global{Request: config.Request{}.WithURL("abc")}
		if got := cfg.Request.URL; !reflect.DeepEqual(got, &url.URL{}) {
			t.Errorf("exp empty *url.URL, got %v", got)
		}
	})

	t.Run("set parsed url", func(t *testing.T) {
		var (
			rawURL    = "http://benchttp.app?cool=true"
			expURL, _ = url.ParseRequestURI(rawURL)
			gotURL    = config.Request{}.WithURL(rawURL).URL
		)

		if !reflect.DeepEqual(gotURL, expURL) {
			t.Errorf("\nexp %v\ngot %v", expURL, gotURL)
		}
	})
}

func TestRequest_Value(t *testing.T) {
	testcases := []struct {
		label  string
		in     config.Request
		expMsg string
	}{
		{
			label:  "return error if url is empty",
			in:     config.Request{},
			expMsg: "empty url",
		},
		{
			label:  "return error if url is invalid",
			in:     config.Request{URL: &url.URL{Scheme: ""}},
			expMsg: "bad url",
		},
		{
			label:  "return error if NewRequest fails",
			in:     config.Request{Method: "é", URL: &url.URL{Scheme: "http"}},
			expMsg: `net/http: invalid method "é"`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			gotReq, gotErr := tc.in.Value()
			if gotErr == nil {
				t.Fatal("exp error, got nil")
			}

			if gotMsg := gotErr.Error(); gotMsg != tc.expMsg {
				t.Errorf("\nexp %q\ngot %q", tc.expMsg, gotMsg)
			}

			if gotReq != nil {
				t.Errorf("exp nil, got %v", gotReq)
			}
		})
	}

	t.Run("return request with added headers", func(t *testing.T) {
		in := config.Request{
			Method: "POST",
			Header: http.Header{"key": []string{"val"}},
			Body:   config.Body{Content: []byte("abc")},
		}.WithURL("http://a.b")

		expReq, err := http.NewRequest(
			in.Method,
			in.URL.String(),
			bytes.NewReader(in.Body.Content),
		)
		if err != nil {
			t.Fatal(err)
		}
		expReq.Header = in.Header

		gotReq, gotErr := in.Value()
		if gotErr != nil {
			t.Fatal(err)
		}

		if !sameRequests(gotReq, expReq) {
			t.Errorf("\nexp %#v\ngot %#v", expReq, gotReq)
		}
	})
}

// helpers

// findErrorOrFail fails t if no error in src matches msg.
func findErrorOrFail(t *testing.T, src []error, msg string) {
	t.Helper()
	for _, err := range src {
		if err.Error() == msg {
			return
		}
	}
	t.Errorf("missing error: %v", msg)
}

func sameRequests(a, b *http.Request) bool {
	if a == nil || b == nil {
		return a == b
	}

	ab, _ := io.ReadAll(a.Body)
	bb, _ := io.ReadAll(b.Body)

	return a.Method == b.Method &&
		a.URL.String() == b.URL.String() &&
		bytes.Equal(ab, bb) &&
		reflect.DeepEqual(a.Header, b.Header)
}
