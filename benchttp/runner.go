package benchttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/engine/benchttp/metrics"
	"github.com/benchttp/engine/benchttp/recorder"
	"github.com/benchttp/engine/benchttp/testsuite"
)

type Runner struct {
	Request *http.Request

	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration

	Tests []testsuite.Case

	OnProgress func(recorder.Progress)

	recorder *recorder.Recorder
}

// DefaultRunner returns a default Runner that is ready to use,
// except for Runner.Request that still needs to be set.
func DefaultRunner() Runner {
	return Runner{
		Concurrency:    10,
		Requests:       100,
		Interval:       0 * time.Second,
		RequestTimeout: 5 * time.Second,
		GlobalTimeout:  30 * time.Second,
	}
}

// WithRequest attaches the given HTTP request to the Runner.
func (r Runner) WithRequest(req *http.Request) Runner {
	r.Request = req
	return r
}

// WithNewRequest calls http.NewRequest with the given parameters
// and attaches the result to the Runner. If the call to http.NewRequest
// returns a non-nil error, it panics with the content of that error.
func (r Runner) WithNewRequest(method, uri string, body io.Reader) Runner {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		panic(err)
	}
	return r.WithRequest(req)
}

func (r Runner) Run(ctx context.Context) (*Report, error) {
	// Validate input config
	if err := r.Validate(); err != nil {
		return nil, err
	}

	// Create and attach request recorder
	r.recorder = recorder.New(r.recorderConfig())

	startTime := time.Now()

	// Run request recorder
	records, err := r.recorder.Record(ctx, r.Request)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)

	agg := metrics.NewAggregate(records)

	testResults := testsuite.Run(agg, r.Tests)

	return newReport(r, duration, agg, testResults), nil
}

// recorderConfig returns a runner.RequesterConfig generated from cfg.
func (r Runner) recorderConfig() recorder.Config {
	return recorder.Config{
		Requests:       r.Requests,
		Concurrency:    r.Concurrency,
		Interval:       r.Interval,
		RequestTimeout: r.RequestTimeout,
		GlobalTimeout:  r.GlobalTimeout,
		OnProgress:     r.OnProgress,
	}
}

// Validate returns a non-nil InvalidConfigError if any of its fields
// does not meet the requirements.
func (r Runner) Validate() error { //nolint:gocognit
	errs := []error{}
	appendError := func(err error) {
		errs = append(errs, err)
	}

	if r.Request == nil {
		appendError(errors.New("Runner.Request must not be nil"))
	}

	if r.Requests < 1 && r.Requests != -1 {
		appendError(fmt.Errorf("requests (%d): want >= 0", r.Requests))
	}

	if r.Concurrency < 1 || r.Concurrency > r.Requests {
		appendError(fmt.Errorf(
			"concurrency (%d): want > 0 and <= requests (%d)",
			r.Concurrency, r.Requests,
		))
	}

	if r.Interval < 0 {
		appendError(fmt.Errorf("interval (%d): want >= 0", r.Interval))
	}

	if r.RequestTimeout < 1 {
		appendError(fmt.Errorf("requestTimeout (%d): want > 0", r.RequestTimeout))
	}

	if r.GlobalTimeout < 1 {
		appendError(fmt.Errorf("globalTimeout (%d): want > 0", r.GlobalTimeout))
	}

	if len(errs) > 0 {
		return &InvalidRunnerError{errs}
	}

	return nil
}
