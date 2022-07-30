package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/runner/internal/metrics"
)

// Report represents a run result as exported by the runner.
type Report struct {
	Metrics  metrics.Aggregate
	Metadata Metadata

	errTemplateFailTriggered error
}

// Metadata contains contextual information about a run.
type Metadata struct {
	Config        Config
	FinishedAt    time.Time
	TotalDuration time.Duration
}

// newReport returns an initialized *Report.
func newReport(m metrics.Aggregate, cfg Config, d time.Duration) *Report {
	return &Report{
		Metrics: m,
		Metadata: Metadata{
			Config:        cfg,
			FinishedAt:    time.Now(), // TODO: change, unreliable
			TotalDuration: d,
		},
	}
}

// String returns a default summary of the Report as a string.
func (rep *Report) String() string {
	var b strings.Builder

	s, err := rep.applyTemplate(rep.Metadata.Config.Output.Template)
	switch {
	case err == nil:
		// template is non-empty and correctly executed,
		// return its result instead of default summary.
		return s
	case errors.Is(err, errTemplateSyntax):
		// template is non-empty but has syntax errors,
		// inform the user about it and fallback to default summary.
		b.WriteString(err.Error())
		b.WriteString("\nFalling back to default summary:\n")
	case errors.Is(err, errTemplateEmpty):
		// template is empty, use default summary.
	}

	rep.writeDefaultSummary(&b)
	return b.String()
}

func (rep *Report) Write(w io.Writer) (int, error) {
	return w.Write([]byte(rep.String()))
}

func (rep *Report) WriteJSON(w io.Writer) (int, error) {
	b, err := json.Marshal(rep)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}

func (rep *Report) writeDefaultSummary(w io.StringWriter) {
	line := func(name string, value interface{}) string {
		const template = "%-18s %v\n"
		return fmt.Sprintf(template, name, value)
	}

	msString := func(d time.Duration) string {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	formatRequests := func(n, max int) string {
		maxString := strconv.Itoa(max)
		if maxString == "-1" {
			maxString = "∞"
		}
		return fmt.Sprintf("%d/%s", n, maxString)
	}

	var (
		m   = rep.Metrics
		cfg = rep.Metadata.Config
	)

	w.WriteString(line("Endpoint", cfg.Request.URL))
	w.WriteString(line("Requests", formatRequests(m.TotalCount, cfg.Runner.Requests)))
	w.WriteString(line("Errors", m.FailureCount))
	w.WriteString(line("Min response time", msString(m.Min)))
	w.WriteString(line("Max response time", msString(m.Max)))
	w.WriteString(line("Mean response time", msString(m.Avg)))
	w.WriteString(line("Total duration", msString(rep.Metadata.TotalDuration)))
}
