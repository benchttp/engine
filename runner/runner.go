package runner

import (
	"context"

	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/output"
	"github.com/benchttp/engine/runner/internal/recorder"
)

type (
	ConfigBody    = config.Body
	ConfigGlobal  = config.Global
	ConfigRequest = config.Request
	ConfigRunner  = config.Runner
	ConfigOutput  = config.Output

	RecorderProgress = recorder.Progress
	RecorderStatus   = recorder.Status

	OutputReport = output.Report
)

const (
	RecorderStatusRunning  = recorder.StatusRunning
	RecorderStatusCanceled = recorder.StatusCanceled
	RecorderStatusTimeout  = recorder.StatusTimeout
	RecorderStatusDone     = recorder.StatusDone

	ConfigFieldMethod         = config.FieldMethod
	ConfigFieldURL            = config.FieldURL
	ConfigFieldHeader         = config.FieldHeader
	ConfigFieldBody           = config.FieldBody
	ConfigFieldRequests       = config.FieldRequests
	ConfigFieldConcurrency    = config.FieldConcurrency
	ConfigFieldInterval       = config.FieldInterval
	ConfigFieldRequestTimeout = config.FieldRequestTimeout
	ConfigFieldGlobalTimeout  = config.FieldGlobalTimeout
	ConfigFieldSilent         = config.FieldSilent
	ConfigFieldTemplate       = config.FieldTemplate
)

var (
	ConfigDefault     = config.Default
	ConfigFieldsUsage = config.FieldsUsage
	ConfigNewBody     = config.NewBody
	ConfigIsField     = config.IsField
)

type Runner struct {
	recorder      *recorder.Recorder
	onStateUpdate func(RecorderProgress)
}

func New(onStateUpdate func(RecorderProgress)) *Runner {
	return &Runner{onStateUpdate: onStateUpdate}
}

func (r *Runner) Run(
	ctx context.Context,
	cfg config.Global,
) (*OutputReport, error) {
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
	r.recorder = recorder.New(recorderConfig(cfg, r.onStateUpdate))

	// Run request recorder
	bk, err := r.recorder.Record(ctx, rq)
	if err != nil {
		return nil, err
	}

	// TODO: compute stats

	return output.New(bk, cfg), nil
}

// Progress returns the current progress of the recording.
// r.Run must have been called before, otherwise it returns
// a zero RecorderProgress.
func (r *Runner) Progress() RecorderProgress {
	if r.recorder == nil {
		return RecorderProgress{}
	}
	return r.recorder.Progress()
}

// recorderConfig returns a runner.RequesterConfig generated from cfg.
func recorderConfig(cfg config.Global, onStateUpdate func(recorder.Progress)) recorder.Config {
	return recorder.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		OnStateUpdate:  onStateUpdate,
	}
}
