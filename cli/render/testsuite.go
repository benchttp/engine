package render

import (
	"io"
	"strings"

	"github.com/benchttp/engine/benchttp/testsuite"

	"github.com/benchttp/engine/cli/render/ansi"
)

func TestSuite(w io.Writer, result testsuite.Result) (int, error) {
	return w.Write([]byte(TestSuiteString(result)))
}

// String returns a default summary of the Report as a string.
func TestSuiteString(result testsuite.Result) string {
	if len(result.OfCases) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(ansi.Bold("→ Test suite"))
	b.WriteString("\n")

	writeResultString(&b, result.Pass)
	b.WriteString("\n")

	for _, tr := range result.OfCases {
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
