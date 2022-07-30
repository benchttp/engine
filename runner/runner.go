package runner

import (
	"context"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/recorder"
)

type (
	RecordingProgress = recorder.Progress
	RecordingStatus   = recorder.Status
)

const (
	StatusRunning  = recorder.StatusRunning
	StatusCanceled = recorder.StatusCanceled
	StatusTimeout  = recorder.StatusTimeout
	StatusDone     = recorder.StatusDone
)

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

	// Generate http request from input config
	rq, err := cfg.Request.Value()
	if err != nil {
		return nil, err
	}

	// Create and attach request recorder
	r.recorder = recorder.New(recorderConfig(cfg, r.onRecordingProgress))

	startTime := time.Now()

	// Run request recorder
	records, err := r.recorder.Record(ctx, rq)
	if err != nil {
		return nil, err
	}

	agg := metrics.Compute(records)

	duration := time.Since(startTime)

	return newReport(agg, cfg, duration), nil
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
