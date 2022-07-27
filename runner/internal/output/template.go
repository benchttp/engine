package output

import (
	"errors"
	"fmt"
	"strings"
	"text/template"
)

var (
	// ErrTemplateFailTriggered a fail triggered by a user
	// using the function {{ fail }} in an output template.
	ErrTemplateFailTriggered = errors.New("test failed")

	errTemplateEmpty  = errors.New("empty template")
	errTemplateSyntax = errors.New("template syntax error")
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
