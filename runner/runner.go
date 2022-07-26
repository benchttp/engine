package runner

import (
	"context"

	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/output"
	"github.com/benchttp/engine/runner/internal/requester"
)

type (
	ConfigBody    = config.Body
	ConfigGlobal  = config.Global
	ConfigRequest = config.Request
	ConfigRunner  = config.Runner
	ConfigOutput  = config.Output

	Requester       = requester.Requester
	RequesterConfig = requester.Config
	RequesterState  = requester.State
	RequesterStatus = requester.Status

	OutputReport = output.Report
)

const (
	RequesterStatusRunning  = requester.StatusRunning
	RequesterStatusCanceled = requester.StatusCanceled
	RequesterStatusTimeout  = requester.StatusTimeout
	RequesterStatusDone     = requester.StatusDone

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

	NewRequester = requester.New

	NewOutput = output.New
)

type Runner struct {
	requester     *Requester
	onStateUpdate func(RequesterState)
}

func New(onStateUpdate func(RequesterState)) *Runner {
	return &Runner{onStateUpdate: onStateUpdate}
}

func (r *Runner) RequesterState() RequesterState {
	return r.requester.State()
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
	r.requester = requester.New(requesterConfig(cfg, r.onStateUpdate))

	// Run request recorder
	bk, err := r.requester.Run(ctx, rq)
	if err != nil {
		return nil, err
	}

	// TODO: compute stats

	return output.New(bk, cfg), nil
}

// requesterConfig returns a runner.RequesterConfig generated from cfg.
func requesterConfig(cfg config.Global, onStateUpdate func(requester.State)) requester.Config {
	return RequesterConfig{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		OnStateUpdate:  onStateUpdate,
	}
}
