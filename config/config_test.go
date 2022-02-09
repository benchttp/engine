package config_test

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/benchttp/runner/config"
)

func TestValidate(t *testing.T) {
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

func TestWithURL(t *testing.T) {
	t.Run("set empty url if invalid", func(t *testing.T) {
		cfg := config.Config{}.WithURL("abc")
		if got := cfg.Request.URL; !reflect.DeepEqual(got, &url.URL{}) {
			t.Errorf("exp empty *url.URL, got %v", got)
		}
	})

	t.Run("set parsed url", func(t *testing.T) {
		var (
			rawURL    = "http://benchttp.app?cool=true"
			expURL, _ = url.ParseRequestURI(rawURL)
			gotURL    = config.Config{}.WithURL(rawURL).Request.URL
		)

		if !reflect.DeepEqual(gotURL, expURL) {
			t.Errorf("\nexp %v\ngot %v", expURL, gotURL)
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
