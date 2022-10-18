package runner

import (
	"context"
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
	recorder            *recorder.Recorder
	onRecordingProgress func(RecordingProgress)
}

func New(onRecordingProgress func(RecordingProgress)) *Runner {
	return &Runner{onRecordingProgress: onRecordingProgress}
}

func (r *Runner) Run(ctx context.Context, cfg Config) (*Report, error) {
	// Validate input config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create and attach request recorder
	r.recorder = recorder.New(recorderConfig(cfg, r.onRecordingProgress))

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
func recorderConfig(
	cfg Config,
	onRecordingProgress func(recorder.Progress),
) recorder.Config {
	return recorder.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		OnProgress:     onRecordingProgress,
	}
}
