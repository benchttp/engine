package requester

import "errors"

var (
	// ErrRequest is returned when an invalid *http.Request is generated
	// by the Requester config.
	ErrRequest = errors.New("invalid request")
	// ErrConnection is returned when the Requester fails to connect to
	// the requested URL.
	ErrConnection = errors.New("connection error")
	// ErrReporting is returned when the Requester fails to send the report.
	ErrReporting = errors.New("reporting error")
)
