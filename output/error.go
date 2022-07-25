package output

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidStrategy reports an unknown strategy set.
	ErrInvalidStrategy = errors.New("invalid strategy")

	// ErrTemplateFailTriggered a fail triggered by a user
	// using the function {{ fail }} in an output template.
	ErrTemplateFailTriggered = errors.New("test failed")

	errTemplateEmpty  = errors.New("empty template")
	errTemplateSyntax = errors.New("template syntax error")
)

// ExportError is the error type returned by Report.Export.
type ExportError struct {
	Errors []error
}

// Error returns the joined errors of ExportError as a string.
func (e *ExportError) Error() string {
	const sep = "\n  - "

	var b strings.Builder
	b.WriteString("output:")
	for _, err := range e.Errors {
		b.WriteString(sep)
		b.WriteString(err.Error())
	}

	return b.String()
}

// ExportErrorOf reads err as an ExportError and returns it.
// If err is not an ExportError, it returns an empty *ExportError
// (non nil) for convenience.
func ExportErrorOf(err error) *ExportError {
	var e *ExportError
	if errors.As(err, &e) && e != nil {
		return e
	}
	return &ExportError{}
}
