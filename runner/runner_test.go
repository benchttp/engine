package runner_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/benchttp/engine/runner"
)

func TestRunner_Validate(t *testing.T) {
	t.Run("return nil if config is valid", func(t *testing.T) {
		brunner := runner.Runner{
			Request:        httptest.NewRequest("GET", "https://a.b/#c?d=e&f=g", nil),
			Requests:       5,
			Concurrency:    5,
			Interval:       5,
			RequestTimeout: 5,
			GlobalTimeout:  5,
		}

		if err := brunner.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("return cumulated errors if config is invalid", func(t *testing.T) {
		brunner := runner.Runner{
			Request:        nil,
			Requests:       -5,
			Concurrency:    -5,
			Interval:       -5,
			RequestTimeout: -5,
			GlobalTimeout:  -5,
		}

		err := brunner.Validate()
		if err == nil {
			t.Fatal("invalid configuration considered valid")
		}

		var errInvalid *runner.InvalidRunnerError
		if !errors.As(err, &errInvalid) {
			t.Fatalf("unexpected error type: %T", err)
		}

		errs := errInvalid.Errors
		assertError(t, errs, `unexpected nil request`)
		assertError(t, errs, `requests (-5): want >= 0`)
		assertError(t, errs, `concurrency (-5): want > 0 and <= requests (-5)`)
		assertError(t, errs, `interval (-5): want >= 0`)
		assertError(t, errs, `requestTimeout (-5): want > 0`)
		assertError(t, errs, `globalTimeout (-5): want > 0`)

		t.Logf("got error:\n%v", errInvalid)
	})
}

// helpers

// assertError fails t if no error in src matches msg.
func assertError(t *testing.T, src []error, msg string) {
	t.Helper()
	for _, err := range src {
		if err.Error() == msg {
			return
		}
	}
	t.Errorf("missing error: %v", msg)
}
