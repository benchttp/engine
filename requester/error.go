package requester

import "errors"

var (
	// ErrConnection is returned when the Requester fails to connect to
	// the requested URL.
	ErrConnection = errors.New("connection error")
	// ErrReporting is returned when the Requester fails to send the report.
	ErrReporting = errors.New("reporting error")
	// ErrCanceled is returned when the Requester.Run context is canceled.
	ErrCanceled = errors.New("canceled")
)
