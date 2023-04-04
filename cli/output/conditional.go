package output

import (
	"io"
)

// ConditionalWriter is an io.Writer that wraps an input writer
// and exposes methods to condition its action.
type ConditionalWriter struct {
	Writer io.Writer
	ok     bool
}

// Write writes b only if ConditionalWriter.Mute is false,
// otherwise it is no-op.
func (w ConditionalWriter) Write(b []byte) (int, error) {
	if !w.ok {
		return 0, nil
	}
	return w.Writer.Write(b)
}

// If sets the write condition to v.
func (w ConditionalWriter) If(v bool) ConditionalWriter {
	return ConditionalWriter{
		Writer: w.Writer,
		ok:     v,
	}
}

// ElseIf either keeps the previous write condition if it is true,
// else it sets it to v.
func (w ConditionalWriter) ElseIf(v bool) ConditionalWriter {
	return ConditionalWriter{
		Writer: w.Writer,
		ok:     w.ok || v,
	}
}
