package benchttp

import (
	"time"

	"github.com/benchttp/engine/benchttp/internal/metrics"
	"github.com/benchttp/engine/benchttp/internal/tests"
)

// Report represents a run result as exported by the runner.
type Report struct {
	Metadata Metadata
	Metrics  metrics.Aggregate
	Tests    tests.SuiteResult
}

// Metadata contains contextual information about a run.
type Metadata struct {
	Runner        Runner
	FinishedAt    time.Time
	TotalDuration time.Duration
}

// newReport returns an initialized *Report.
func newReport(
	r Runner,
	d time.Duration,
	m metrics.Aggregate,
	t tests.SuiteResult,
) *Report {
	return &Report{
		Metrics: m,
		Tests:   t,
		Metadata: Metadata{
			Runner:        r,
			FinishedAt:    time.Now(), // TODO: change, unreliable
			TotalDuration: d,
		},
	}
}
