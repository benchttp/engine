package export

import (
	"errors"
	"fmt"
)

var (
	// ErrJSONMarshal reports an error marshaling JSON.
	ErrJSONMarshal = errors.New("export: error marshaling JSON")
	// ErrFileCreate reports an error creating a file.
	ErrFileCreate = errors.New("export: error creating file")
	// ErrFileWrite reports an error writing a file.
	ErrFileWrite = errors.New("export: error writing file")
	// ErrHTTPRequest reports an error generating a HTTP request.
	ErrHTTPRequest = errors.New("export: request error")
	// ErrHTTPConnection reports an error sending a HTTP request.
	ErrHTTPConnection = errors.New("export: HTTP connection error")
	// ErrHTTPResponse reports a HTTP response error, such as bad status code.
	ErrHTTPResponse = &HTTPResponseError{}
)

// HTTPResponseError is an HTTP error due to a bad response code.
// It contains the received HTTP status code.
type HTTPResponseError struct {
	Code int
}

// Error implements error.
func (e *HTTPResponseError) Error() string {
	return fmt.Sprintf("export: HTTP response error: status code: %d", e.Code)
}

// Is returns true if err's type is HTTPResponseError.
// It allows comparison with ErrHTTPResponse using errors.Is:
// 	errors.Is(&HTTPResponseError{}, ErrHTTPResponse) // true
// 	errors.Is(errors.New("any"), ErrHTTPResponse) // false
// It does not perform check equality on the status code.
func (e *HTTPResponseError) Is(err error) bool {
	if err == nil {
		return false
	}
	var errHTTPResponse *HTTPResponseError
	return errors.As(err, &errHTTPResponse)
}

// WithCode returns a *HTTPResponseError initialized with code.
func (e *HTTPResponseError) WithCode(code int) *HTTPResponseError {
	return &HTTPResponseError{Code: code}
}
