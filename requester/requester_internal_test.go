package requester

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/dispatcher"
)

const (
	badURL  = "abc"
	goodURL = "http://a.b"
)

var errTest = errors.New("test-generated error")

func TestRun(t *testing.T) {
	testcases := []struct {
		label string
		req   *Requester
		exp   error
	}{
		{
			label: "return ErrRequest early on request error",
			req:   New(config.New(badURL, -1, 1, 0, 0)),
			exp:   ErrRequest,
		},
		{
			label: "return ErrConnection early on connection error",
			req:   New(config.New(goodURL, -1, 1, 0, 0)),
			exp:   ErrConnection,
		},
		{
			label: "return dispatcher.ErrInvalidValue early on bad dispatcher value",
			req:   withNoopTransport(New(config.New(goodURL, 1, 2, time.Second, time.Minute))),
			exp:   dispatcher.ErrInvalidValue,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			gotRep, gotErr := tc.req.Run()

			if !errors.Is(gotErr, tc.exp) {
				t.Errorf("unexpected error value:\nexp %v\ngot %v", tc.exp, gotErr)
			}

			if !reflect.ValueOf(gotRep).IsZero() {
				t.Errorf("report value:\nexp %v\ngot %v", Report{}, gotRep)
			}
		})
	}

	t.Run("record failing requests", func(t *testing.T) {
		r := withErrTransport(New(config.New(goodURL, 1, 1, time.Second, time.Minute)))

		rep, err := r.Run()
		if err != nil {
			t.Errorf("exp nil error, got %v", err)
		}

		if rep.Length != 1 {
			t.Errorf("unexpected Report.Length: exp 1, got %d", rep.Length)
		}

		if rep.Success != 0 {
			t.Errorf("unexpected Report.Success: exp 0, got %d", rep.Success)
		}

		if rep.Fail != 1 {
			t.Errorf("unexpected Report.Fail: exp 1, got %d", rep.Fail)
		}

		t.Log(rep)
	})

	t.Run("happy path", func(t *testing.T) {
		r := withNoopTransport(New(config.New(goodURL, 1, 1, time.Second, 2*time.Second)))

		rep, err := r.Run()
		if err != nil {
			t.Errorf("exp nil error, got %v", err)
		}

		if rep.Length != 1 {
			t.Errorf("unexpected Report.Length: exp 1, got %d", rep.Length)
		}

		if rep.Success != 1 {
			t.Errorf("unexpected Report.Success: exp 1, got %d", rep.Success)
		}

		if rep.Fail != 0 {
			t.Errorf("unexpected Report.Fail: exp 0, got %d", rep.Fail)
		}

		t.Log(rep)
	})
}

// helpers

type noopTransport struct{}

func (noopTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}

func withNoopTransport(req *Requester) *Requester {
	req.client.Transport = noopTransport{}
	return req
}

type errTransport struct{}

func (errTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{Body: unreadableReadCloser{}}, nil
}

func withErrTransport(req *Requester) *Requester {
	req.client.Transport = errTransport{}
	return req
}

type unreadableReadCloser struct{}

func (unreadableReadCloser) Read([]byte) (int, error) {
	return 0, errTest
}

func (unreadableReadCloser) Close() error {
	return nil
}
