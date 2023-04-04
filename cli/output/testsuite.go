package output

import (
	"io"
	"strings"

	"github.com/benchttp/engine/benchttp"

	"github.com/benchttp/engine/cli/output/ansi"
)

func TestSuite(w io.Writer, suite benchttp.TestSuiteResults) (int, error) {
	return w.Write([]byte(TestSuiteString(suite)))
}

// String returns a default summary of the Report as a string.
func TestSuiteString(suite benchttp.TestSuiteResults) string {
	if len(suite.Results) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(ansi.Bold("→ Test suite"))
	b.WriteString("\n")

	writeResultString(&b, suite.Pass)
	b.WriteString("\n")

	for _, tr := range suite.Results {
		writeIndent(&b, 1)
		writeResultString(&b, tr.Pass)
		b.WriteString(" ")
		b.WriteString(tr.Input.Name)

		if !tr.Pass {
			b.WriteString("\n ")
			writeIndent(&b, 3)
			b.WriteString(ansi.Bold("→ "))
			b.WriteString(tr.Summary)
		}

		b.WriteString("\n")
	}

	return b.String()
}

func writeResultString(w io.StringWriter, pass bool) {
	if pass {
		w.WriteString(ansi.Green("PASS"))
	} else {
		w.WriteString(ansi.Red("FAIL"))
	}
}

func writeIndent(w io.StringWriter, count int) {
	if count <= 0 {
		return
	}
	const baseIndent = "  "
	w.WriteString(strings.Repeat(baseIndent, count))
}
