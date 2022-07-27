package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/internal/cli/ansi"
	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/tests"
)

// Report represents a run result as exported by the runner.
type Report struct {
	Metrics  metrics.Aggregate
	Metadata Metadata
	Tests    tests.SuiteResult

	errTemplateFailTriggered error
}

// Metadata contains contextual information about a run.
type Metadata struct {
	Config        config.Global
	FinishedAt    time.Time
	TotalDuration time.Duration
}

// New returns an initialized *Report.
func New(
	m metrics.Aggregate,
	cfg config.Global,
	d time.Duration,
	testResults tests.SuiteResult,
) *Report {
	return &Report{
		Metrics: m,
		Tests:   testResults,
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
		rep.writeTestsResult(&b)
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
	rep.writeTestsResult(&b)

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

	w.WriteString(ansi.Bold("→ Summary"))
	w.WriteString("\n")
	w.WriteString(line("Endpoint", cfg.Request.URL))
	w.WriteString(line("Requests", formatRequests(m.TotalCount, cfg.Runner.Requests)))
	w.WriteString(line("Errors", m.FailureCount))
	w.WriteString(line("Min response time", msString(m.Min)))
	w.WriteString(line("Max response time", msString(m.Max)))
	w.WriteString(line("Mean response time", msString(m.Avg)))
	w.WriteString(line("Total duration", msString(rep.Metadata.TotalDuration)))
}

func (rep *Report) writeTestsResult(w io.StringWriter) {
	sr := rep.Tests
	if len(sr.Results) == 0 {
		return
	}

	w.WriteString("\n")
	w.WriteString(ansi.Bold("→ Test suite"))
	w.WriteString("\n")

	writeResultString(w, sr.Pass)
	w.WriteString("\n")

	for _, tr := range sr.Results {
		writeIndent(w, 1)
		writeResultString(w, tr.Pass)
		w.WriteString(": ")
		w.WriteString(tr.Name)

		if !tr.Pass {
			w.WriteString("\n")
			writeIndent(w, 2)
			w.WriteString(tr.Explain)
		}

		w.WriteString("\n")
	}
}

func writeResultString(w io.StringWriter, pass bool) {
	if pass {
		w.WriteString(ansi.Green("√"))
		w.WriteString(" PASS")
	} else {
		w.WriteString(ansi.Red("x"))
		w.WriteString(" FAIL")
	}
}

func writeIndent(w io.StringWriter, count int) {
	if count <= 0 {
		return
	}
	const baseIndent = "  "
	w.WriteString(strings.Repeat(baseIndent, count))
}
