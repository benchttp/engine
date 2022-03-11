package config

import "strings"

// InvalidConfigError is the errors returned by Global.Validate
// when values are missing or invalid.
type InvalidConfigError struct {
	Errors []error
}

// Error returns the joined errors of InvalidConfigError as a string.
func (e *InvalidConfigError) Error() string {
	const sep = "\n  - "

	var b strings.Builder

	b.WriteString("Invalid value(s) provided:")
	for _, err := range e.Errors {
		if err != nil {
			b.WriteString(sep)
			b.WriteString(err.Error())
		}
	}
	return b.String()
}
