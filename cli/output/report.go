package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/benchttp"

	"github.com/benchttp/engine/cli/output/ansi"
)

func ReportSummary(w io.Writer, rep *benchttp.Report) (int, error) {
	return w.Write([]byte(ReportSummaryString(rep)))
}

// String returns a default summary of the Report as a string.
func ReportSummaryString(rep *benchttp.Report) string {
	var b strings.Builder

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

	m := rep.Metrics
	r := rep.Metadata.Runner

	b.WriteString(ansi.Bold("→ Summary"))
	b.WriteString("\n")
	b.WriteString(line("Endpoint", r.Request.URL))
	b.WriteString(line("Requests", formatRequests(len(m.Records), r.Requests)))
	b.WriteString(line("Errors", len(m.RequestFailures)))
	b.WriteString(line("Min response time", msString(m.ResponseTimes.Min)))
	b.WriteString(line("Max response time", msString(m.ResponseTimes.Max)))
	b.WriteString(line("Mean response time", msString(m.ResponseTimes.Mean)))
	b.WriteString(line("Total duration", msString(rep.Metadata.TotalDuration)))
	b.WriteString("\n")

	return b.String()
}
