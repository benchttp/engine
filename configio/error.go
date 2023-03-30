package configio

import (
	"errors"
	"fmt"
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

	// ErrFileParse signals an error parsing a retrieved config file.
	// It is returned if it contains an unexpected field or an unexpected
	// value for a field.
	ErrFileParse = errors.New("parsing error: invalid config file")

	// ErrFileCircular signals a circular reference in the config file.
	ErrFileCircular = errors.New("circular reference detected")
)

func panicInternal(funcname, detail string) {
	const reportURL = "https://github.com/benchttp/engine/issues/new"
	source := fmt.Sprintf("configio.%s", funcname)
	panic(fmt.Sprintf(
		"%s: unexpected internal error: %s, please file an issue at %s",
		source, detail, reportURL,
	))
}
