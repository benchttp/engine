package file

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrFileNotFound signals a config file not found.
	ErrFileNotFound = errors.New("file not found")

	// ErrFileRead signals an error trying to read a config file.
	// It can be due to a corrupted file or an invalid permission
	// for instance.
	ErrFileRead = errors.New("invalid file")

	// ErrFileExt signals an unsupported extension for the config file.
	ErrFileExt = errors.New("invalid extension")

	// ErrParse signals an error parsing a retrieved config file.
	// It is returned if it contains an unexpected field or an unexpected
	// value for a field.
	ErrParse = errors.New("parsing error: invalid config file")
)

// errWithDetails returns an error wrapping err, appended with a string
// representation of details separated by ": ".
func errWithDetails(err error, details ...interface{}) error {
	detailsStr := make([]string, len(details))
	for i := range details {
		detailsStr[i] = fmt.Sprint(details[i])
	}
	return fmt.Errorf("%w: %s", err, strings.Join(detailsStr, ": "))
}
