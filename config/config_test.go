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

func TestOverride(t *testing.T) {
	t.Run("do not override unspecified fields", func(t *testing.T) {
		baseCfg := config.Config{}
		newCfg := config.New("http://a.b?p=2", 1, 2, 3, 4)

		if gotCfg := baseCfg.Override(newCfg); gotCfg != baseCfg {
			t.Errorf("overrode unexpected fields:\nexp %#v\ngot %#v", baseCfg, gotCfg)
		}
	})

	t.Run("override specified fields", func(t *testing.T) {
		baseCfg := config.Config{}
		newCfg := config.New("http://a.b?p=2", 1, 2, 3, 4)
		fields := []string{
			config.FieldMethod,
			config.FieldURL,
			config.FieldTimeout,
			config.FieldRequests,
			config.FieldConcurrency,
			config.FieldGlobalTimeout,
		}

		if gotCfg := baseCfg.Override(newCfg, fields...); !reflect.DeepEqual(gotCfg, newCfg) {
			t.Errorf("did not override expected fields:\nexp %v\ngot %v", baseCfg, gotCfg)
			t.Log(fields)
		}
	})
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
