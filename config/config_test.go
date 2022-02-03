package config_test

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
)

func TestConfigValidation(t *testing.T) {
	t.Run("test valid configuration", func(t *testing.T) {
		cfg := config.New("https://github.com/benchttp/", 5, 5, 5, 5)
		err := cfg.Validate()
		if err != nil {
			t.Errorf("valid configuration not considered as such")
		}
	})

	t.Run("test invalid configuration returns ErrInvalid error with correct messages", func(t *testing.T) {
		cfg := config.New("github-com/benchttp/", -5, -5, -5, -5)
		err := cfg.Validate()
		if err == nil {
			t.Errorf("invalid configuration considered valid")
		} else {
			if !errorContains(err, "-url: "+cfg.Request.URL.String()+" is not a valid url") {
				t.Errorf("\n- information about invalid url missing from error message")
			}
			if !errorContains(err, "-requests: must be >= 0, we got ") {
				t.Errorf("\n- information about invalid requests number missing from error message")
			}
			if !errorContains(err, "-concurrency: must be > 0, we got ") {
				t.Errorf("\n- information about invalid concurrency number missing from error message")
			}
			if !errorContains(err, "-timeout: must be > 0, we got") {
				t.Errorf("\n- information about invalid timeout missing from error message")
			}
			if !errorContains(err, "-globalTimeout: must be > 0, we got ") {
				t.Errorf("\n- information about invalid globalTimeout missing from error message")
			}
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("do not override with zero values", func(t *testing.T) {
		cfgBase := newConfig()
		cfgZero := config.Config{}

		if got := config.Merge(cfgBase, cfgZero); !reflect.DeepEqual(got, cfgBase) {
			t.Errorf("overrode with zero values:\nexp %#v\ngot %#v", cfgBase, got)
		}
	})

	t.Run("override with non-zero values", func(t *testing.T) {
		cfgBase := newConfig()
		cfgOver := config.Config{
			Request: config.Request{
				Method: "POST",
				URL: &url.URL{
					Host: "example",
				},
				Timeout: 2 * time.Second,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      2,
				Concurrency:   2,
				GlobalTimeout: 2 * time.Second,
			},
		}

		if got := config.Merge(cfgBase, cfgOver); !reflect.DeepEqual(got, cfgOver) {
			t.Errorf(
				"did not override with non-zero values\nexp %#v\ngot %#v",
				cfgOver, got,
			)
		}
	})

	t.Run("override with non-zero values selectively", func(t *testing.T) {
		cfgBase := newConfig()
		cfgOver := config.Config{}
		cfgOver.Request.Method = "POST"
		cfgOver.RunnerOptions.Concurrency = 10

		exp := config.Config{
			Request: config.Request{
				Method:  cfgOver.Request.Method,
				URL:     cfgBase.Request.URL,
				Timeout: cfgBase.Request.Timeout,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      cfgBase.RunnerOptions.Requests,
				Concurrency:   cfgOver.RunnerOptions.Concurrency,
				GlobalTimeout: cfgBase.RunnerOptions.GlobalTimeout,
			},
		}

		if got := config.Merge(cfgBase, cfgOver); got != exp {
			t.Errorf(
				"did not selectively override with non-zero values\nexp %#v\ngot %#v",
				exp, got,
			)
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("zero value params return empty config", func(t *testing.T) {
		exp := config.Config{Request: config.Request{URL: &url.URL{}}}
		if got := config.New("", 0, 0, 0, 0); !reflect.DeepEqual(got, exp) {
			t.Errorf("returned non-zero config:\nexp %#v\ngot %#v", exp, got)
		}
	})

	t.Run("non-zero params return initialized config", func(t *testing.T) {
		var (
			rawURL      = "http://example.com"
			urlURL, _   = url.ParseRequestURI(rawURL)
			requests    = 1
			concurrency = 2
			reqTimeout  = 3 * time.Second
			glbTimeout  = 4 * time.Second
		)

		exp := config.Config{
			Request: config.Request{
				Method:  "",
				URL:     urlURL,
				Timeout: reqTimeout,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      requests,
				Concurrency:   concurrency,
				GlobalTimeout: glbTimeout,
			},
		}

		got := config.New(rawURL, requests, concurrency, reqTimeout, glbTimeout)

		if !reflect.DeepEqual(got, exp) {
			t.Errorf("returned unexpected config:\nexp %#v\ngot %#v", exp, got)
		}
	})
}

// helpers

func newConfig() config.Config {
	return config.Config{
		Request: config.Request{
			Method: "GET",
			URL: &url.URL{
				Host:     "localhost",
				RawQuery: "delay=200ms",
			},
			Timeout: 1 * time.Second,
		},
		RunnerOptions: config.RunnerOptions{
			Requests:      1,
			Concurrency:   1,
			GlobalTimeout: 1 * time.Second,
		},
	}
}

// To check that the error message is as expected
func errorContains(err error, expected string) bool {
	if err == nil {
		return expected == ""
	}
	if expected == "" {
		return false
	}
	return strings.Contains(err.Error(), expected)
}
