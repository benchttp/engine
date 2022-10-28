package recorder

import (
	"errors"
	"fmt"
)

var (
	// ErrConnection is returned when the Recorder fails to connect to
	// the requested URL.
	ErrConnection = errors.New("connection error")
	// ErrCanceled is returned when the Recorder.Run context is canceled.
	ErrCanceled = errors.New("canceled")
)

// recordErr wraps and returns err as a string, marking it as an error
// that happened when recording the request.
func recordErr(err error) string {
	return fmt.Sprintf("recording error: %s", err)
}
