package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/internal/cli/ansi"
	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/metrics"
)

type basicStats struct {
	Min, Max, Mean time.Duration
}

func (s basicStats) isZero() bool {
	return s == basicStats{}
}

// Report represent a benchmark result as exported by the runner.
type Report struct {
	Metrics  metrics.Aggregate
	Metadata Metadata

	errTemplateFailTriggered error

	log func(v ...interface{})
}

type Metadata struct {
	Config        config.Global
	FinishedAt    time.Time
	TotalDuration time.Duration
}

// New returns a Report initialized with the input benchmark and the config
// used to run it.
func New(m metrics.Aggregate, cfg config.Global, d time.Duration) *Report {
	outputLogger := newLogger(cfg.Output.Silent)
	return &Report{
		Metrics: m,
		Metadata: Metadata{
			Config:        cfg,
			FinishedAt:    time.Now(), // TODO: change, unreliable
			TotalDuration: d,
		},
		log: outputLogger.Println,
	}
}

// newLogger returns the logger to be used by Report.
func newLogger(silent bool) *log.Logger {
	var w io.Writer = os.Stdout
	if silent {
		w = nopWriter{}
	}
	return log.New(w, ansi.Bold("→ "), 0)
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

	// generate default summary

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

	b.WriteString(line("Endpoint", cfg.Request.URL))
	b.WriteString(line("Requests", formatRequests(m.TotalCount, cfg.Runner.Requests)))
	b.WriteString(line("Errors", m.FailureCount))
	b.WriteString(line("Min response time", msString(m.Min)))
	b.WriteString(line("Max response time", msString(m.Max)))
	b.WriteString(line("Mean response time", msString(m.Avg)))
	b.WriteString(line("Total duration", msString(rep.Metadata.TotalDuration)))
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

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return 0, nil }
