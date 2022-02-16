package requester

import "errors"

var (
	// ErrConnection is returned when the Requester fails to connect to
	// the requested URL.
	ErrConnection = errors.New("connection error")
	// ErrReporting is returned when the Requester fails to send the report.
	ErrReporting = errors.New("reporting error")
	// ErrRequestBody is returned when the Requester fails to get the body of the request to copy it
	ErrRequestBody = errors.New("request body error")
)
