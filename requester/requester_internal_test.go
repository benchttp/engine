package requester

import (
	"bytes"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/benchttp/runner/dispatcher"
)

var errTest = errors.New("test-generated error")

func TestRun(t *testing.T) {
	testcases := []struct {
		label string
		req   *Requester
		exp   error
	}{
		{
			label: "return ErrConnection early on connection error",
			req: New(Config{
				Requests:       -1,
				Concurrency:    1,
				RequestTimeout: 1 * time.Second,
				GlobalTimeout:  0,
			},
			),
			exp: ErrConnection,
		},
		{
			label: "return dispatcher.ErrInvalidValue early on bad dispatcher value",
			req: withNoopTransport(New(Config{
				Requests:       1,
				Concurrency:    2, // bad: Concurrency > Requests
				RequestTimeout: 1 * time.Second,
				GlobalTimeout:  3 * time.Second,
			},
			)),
			exp: dispatcher.ErrInvalidValue,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			gotRep, gotErr := tc.req.Run(validRequest())

			if !errors.Is(gotErr, tc.exp) {
				t.Errorf("unexpected error value:\nexp %v\ngot %v", tc.exp, gotErr)
			}

			if !reflect.ValueOf(gotRep).IsZero() {
				t.Errorf("report value:\nexp %v\ngot %v", Report{}, gotRep)
			}
		})
	}

	t.Run("record failing requests", func(t *testing.T) {
		r := withErrTransport(New(Config{
			Requests:       1,
			Concurrency:    1,
			RequestTimeout: 1 * time.Second,
			GlobalTimeout:  3 * time.Second,
		}))

		rep, err := r.Run(validRequest())
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
		r := withNoopTransport(New(Config{
			Requests:       1,
			Concurrency:    1,
			RequestTimeout: 1 * time.Second,
			GlobalTimeout:  3 * time.Second,
		}))

		rep, err := r.Run(validRequestWithBody([]byte(`{"key0": "val0", "key1": "val1"}`)))
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

	t.Run("use interval", func(t *testing.T) {
		const (
			requests    = 12
			concurrency = 3
			interval    = 30 * time.Millisecond
			baseMargin  = 4 * time.Millisecond
		)

		var (
			mu             sync.Mutex
			start          time.Time
			currentRequest int

			expTimes = []time.Duration{
				0 * time.Millisecond, 0 * time.Millisecond, 0 * time.Millisecond,
				30 * time.Millisecond, 30 * time.Millisecond, 30 * time.Millisecond,
				60 * time.Millisecond, 60 * time.Millisecond, 60 * time.Millisecond,
				90 * time.Millisecond, 90 * time.Millisecond, 90 * time.Millisecond,
			}
			gotTimes = make([]time.Duration, 0, requests)
		)

		cfg := Config{
			Concurrency:   concurrency,
			Requests:      requests,
			Interval:      interval,
			GlobalTimeout: 5 * time.Second,
		}

		r := withCallbackTransport(New(cfg), func() {
			defer func() {
				currentRequest++
				mu.Unlock()
			}()
			mu.Lock()

			// ignore first request from r.ping()
			if currentRequest == 0 {
				return
			}

			// actual first request: start timer now
			if currentRequest == 1 {
				start = time.Now()
			}

			elapsed := time.Since(start)
			gotTimes = append(gotTimes, elapsed)
		})

		if _, err := r.Run(validRequest()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(gotTimes) != len(expTimes) {
			t.Logf("exp %v\ngot %v", expTimes, gotTimes)
			t.Fatal("unexpected requests count")
		}

		fail := false
		for i, gotTime := range gotTimes {
			// Delay accumulates each non-concurrent iteration, so we increase
			// the margin accordingly each step.
			// With baseMargin == 4ms and concurrency == 3, this empirically
			// determined formula gives the following margins:
			// 	4ms, 4ms, 4ms    // 0ms <= d <= 4ms
			// 	8ms, 8ms, 8ms    // 30ms <= d <= 38ms
			// 	12ms, 12ms, 12ms // 60ms <= d <= 72ms
			// 	16ms, 16ms, 16ms // 90ms <= d <= 106ms
			margin := baseMargin + time.Duration(i/concurrency)*baseMargin
			if gotTime < expTimes[i] || gotTime > expTimes[i]+margin {
				fail = true
			}
		}

		if fail {
			t.Errorf("unexpected interval:\nexp %v\ngot %v", expTimes, gotTimes)
		}
	})
}

// helpers

type callbackTransport struct{ callback func() }

func (t callbackTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.callback()
	return &http.Response{}, nil
}

func withCallbackTransport(req *Requester, callback func()) *Requester {
	req.newTransport = func() http.RoundTripper {
		return callbackTransport{callback: callback}
	}
	return req
}

func withNoopTransport(req *Requester) *Requester {
	return withCallbackTransport(req, func() {})
}

type errTransport struct{}

func (errTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{Body: unreadableReadCloser{}}, nil
}

func withErrTransport(req *Requester) *Requester {
	req.newTransport = func() http.RoundTripper {
		return errTransport{}
	}
	return req
}

type unreadableReadCloser struct{}

func (unreadableReadCloser) Read([]byte) (int, error) {
	return 0, errTest
}

func (unreadableReadCloser) Close() error {
	return nil
}

const validURI = "http://a.b"

func validRequest() *http.Request {
	request, _ := http.NewRequest("", validURI, nil)
	return request
}

func validRequestWithBody(body []byte) *http.Request {
	request, _ := http.NewRequest("POST", validURI, bytes.NewReader(body))
	return request
}
