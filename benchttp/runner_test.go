package benchttp_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/benchttp/sdk/benchttp"
)

func TestRunner_WithRequest(t *testing.T) {
	request := httptest.NewRequest("POST", "https://example.com", nil)
	runner := benchttp.DefaultRunner().WithRequest(request)

	if runner.Request != request {
		t.Error("request not set")
	}
}

func TestRunner_WithNewRequest(t *testing.T) {
	t.Run("attach valid request", func(t *testing.T) {
		const method = "POST"
		const urlString = "https://example.com"

		runner := benchttp.DefaultRunner().WithNewRequest(method, urlString, nil)

		if runner.Request.Method != method || runner.Request.URL.String() != urlString {
			t.Error("request incorrectly seet")
		}
	})

	t.Run("panic for bad request params", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("did not panic")
			}
		}()
		_ = benchttp.DefaultRunner().WithNewRequest("ù", "", nil)
	})
}

func TestRunner_Validate(t *testing.T) {
	t.Run("return nil if config is valid", func(t *testing.T) {
		runner := benchttp.Runner{
			Request:        httptest.NewRequest("GET", "https://a.b/#c?d=e&f=g", nil),
			Requests:       5,
			Concurrency:    5,
			Interval:       5,
			RequestTimeout: 5,
			GlobalTimeout:  5,
		}

		if err := runner.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("return cumulated errors if config is invalid", func(t *testing.T) {
		runner := benchttp.Runner{
			Request:        nil,
			Requests:       -5,
			Concurrency:    -5,
			Interval:       -5,
			RequestTimeout: -5,
			GlobalTimeout:  -5,
		}

		err := runner.Validate()
		if err == nil {
			t.Fatal("invalid configuration considered valid")
		}

		var errInvalid *benchttp.InvalidRunnerError
		if !errors.As(err, &errInvalid) {
			t.Fatalf("unexpected error type: %T", err)
		}

		errs := errInvalid.Errors
		assertError(t, errs, "Runner.Request must not be nil")
		assertError(t, errs, "requests (-5): want >= 0")
		assertError(t, errs, "concurrency (-5): want > 0 and <= requests (-5)")
		assertError(t, errs, "interval (-5): want >= 0")
		assertError(t, errs, "requestTimeout (-5): want > 0")
		assertError(t, errs, "globalTimeout (-5): want > 0")

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
