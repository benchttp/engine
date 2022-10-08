package report

import (
	"time"

	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/tests"
)

// Report represents a run result as exported by the runner.
type Report struct {
	Metadata Metadata
	Metrics  metrics.Aggregate
	Tests    tests.SuiteResult
}

// Metadata contains contextual information about a run.
type Metadata struct {
	Config        config.Global
	FinishedAt    time.Time
	TotalDuration time.Duration
}

// New returns an initialized *Report.
func New(
	cfg config.Global,
	d time.Duration,
	m metrics.Aggregate,
	t tests.SuiteResult,
) *Report {
	return &Report{
		Metrics: m,
		Tests:   t,
		Metadata: Metadata{
			Config:        cfg,
			FinishedAt:    time.Now(), // TODO: change, unreliable
			TotalDuration: d,
		},
	}
}
