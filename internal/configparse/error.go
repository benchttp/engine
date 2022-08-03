package configparse

import (
	"errors"
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

	// ErrCircularExtends signals a circular reference in the config file.
	ErrCircularExtends = errors.New("circular reference detected")
)
