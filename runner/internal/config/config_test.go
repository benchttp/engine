package config_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/benchttp/engine/runner/internal/config"
)

func TestGlobal_Validate(t *testing.T) {
	t.Run("return nil if config is valid", func(t *testing.T) {
		cfg := config.Global{
			Request: validRequest(),
			Runner: config.Runner{
				Requests:       5,
				Concurrency:    5,
				Interval:       5,
				RequestTimeout: 5,
				GlobalTimeout:  5,
			},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("return cumulated errors if config is invalid", func(t *testing.T) {
		cfg := config.Global{
			Request: nil,
			Runner: config.Runner{
				Requests:       -5,
				Concurrency:    -5,
				Interval:       -5,
				RequestTimeout: -5,
				GlobalTimeout:  -5,
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
		findErrorOrFail(t, errs, `unexpected nil request`)
		findErrorOrFail(t, errs, `requests (-5): want >= 0`)
		findErrorOrFail(t, errs, `concurrency (-5): want > 0 and <= requests (-5)`)
		findErrorOrFail(t, errs, `interval (-5): want >= 0`)
		findErrorOrFail(t, errs, `requestTimeout (-5): want > 0`)
		findErrorOrFail(t, errs, `globalTimeout (-5): want > 0`)

		t.Logf("got error:\n%v", errInvalid)
	})
}

// helpers

func validRequest() *http.Request {
	req, err := http.NewRequest("GET", "https://a.b#c?d=e&f=g", nil)
	if err != nil {
		panic(err)
	}
	return req
}

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
