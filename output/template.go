package output

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/benchttp/runner/requester"
)

// applyTemplate applies Report to a template using given pattern and returns
// the result as a string. If pattern == "", it returns errTemplateEmpty.
// If an error occurs parsing the pattern or executing the template,
// it returns errTemplateSyntax.
func (rep *Report) applyTemplate(pattern string) (string, error) {
	if pattern == "" {
		return "", errTemplateEmpty
	}

	t, err := template.
		New("report").
		Funcs(rep.templateFuncs()).
		Parse(pattern)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errTemplateSyntax, err)
	}

	var b strings.Builder
	if err := t.Execute(&b, rep); err != nil {
		return "", fmt.Errorf("%w: %s", errTemplateSyntax, err)
	}

	return b.String(), nil
}

// templateFuncs returns a template.FuncMap defining template functions
// that are specific to the Report: stats, event, fail.
func (rep *Report) templateFuncs() template.FuncMap {
	return template.FuncMap{
		// stats computes basic stats for the Report if not already done,
		// and returns the results as basicStats.
		"stats": func() basicStats {
			if rep.stats.isZero() {
				rep.stats.Min, rep.stats.Max, rep.stats.Mean = rep.Benchmark.Stats()
			}
			return rep.stats
		},

		// event retrieves an event from the input record given a its name
		// and returns its time.
		"event": func(rec requester.Record, name string) time.Duration {
			for _, e := range rec.Events {
				if e.Name == name {
					return e.Time
				}
			}
			return 0
		},

		// fail sets rep.errTplFailTriggered to the given error, causing
		// the test to fail
		"fail": func(a ...interface{}) string {
			if rep.errTemplateFailTriggered == nil {
				rep.errTemplateFailTriggered = fmt.Errorf(
					"%w: %s",
					ErrTemplateFailTriggered, fmt.Sprint(a...),
				)
			}
			return ""
		},
	}
}
