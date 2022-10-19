package runner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/tests"
)

type (
	RecordingProgress = recorder.Progress
	RecordingStatus   = recorder.Status

	MetricsAggregate = metrics.Aggregate
	MetricsField     = metrics.Field
	MetricsValue     = metrics.Value
	MetricsTimeStats = metrics.TimeStats

	TestCase         = tests.Case
	TestPredicate    = tests.Predicate
	TestSuiteResults = tests.SuiteResult
	TestCaseResult   = tests.CaseResult
)

const (
	StatusRunning  = recorder.StatusRunning
	StatusCanceled = recorder.StatusCanceled
	StatusTimeout  = recorder.StatusTimeout
	StatusDone     = recorder.StatusDone
)

var ErrCanceled = recorder.ErrCanceled

type Runner struct {
	Request *http.Request

	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration

	Tests []tests.Case

	recorder            *recorder.Recorder
	onRecordingProgress func(RecordingProgress)
}

func New(onRecordingProgress func(RecordingProgress)) *Runner {
	return &Runner{onRecordingProgress: onRecordingProgress}
}

func (r *Runner) Run(ctx context.Context, cfg Runner) (*Report, error) {
	// Validate input config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create and attach request recorder
	r.recorder = recorder.New(r.recorderConfig())

	startTime := time.Now()

	// Run request recorder
	records, err := r.recorder.Record(ctx, cfg.Request)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)

	agg := metrics.NewAggregate(records)

	testResults := tests.Run(agg, cfg.Tests)

	return newReport(cfg, duration, agg, testResults), nil
}

// Progress returns the current progress of the recording.
// r.Run must have been called before, otherwise it returns
// a zero RecorderProgress.
func (r *Runner) Progress() RecordingProgress {
	if r.recorder == nil {
		return RecordingProgress{}
	}
	return r.recorder.Progress()
}

// recorderConfig returns a runner.RequesterConfig generated from cfg.
func (r *Runner) recorderConfig() recorder.Config {
	return recorder.Config{
		Requests:       r.Requests,
		Concurrency:    r.Concurrency,
		Interval:       r.Interval,
		RequestTimeout: r.RequestTimeout,
		GlobalTimeout:  r.GlobalTimeout,
		OnProgress:     r.onRecordingProgress,
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
		appendError(errors.New("unexpected nil request"))
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
