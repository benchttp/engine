package record

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Record is a summary of an http call.
type Record struct {
	Cost  time.Duration
	Code  int
	Bytes int
	Error error
}

// New returns a Record that summarizes the given http response,
// attaching the duration and a non-nil error if any occurs
// in the reading process.
func New(resp *http.Response, t time.Duration) Record {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	r := Record{
		Code:  resp.StatusCode,
		Cost:  t,
		Bytes: len(body),
	}

	if err != nil {
		r.Error = err
	}

	return r
}

// String prints the Record{}.Cost. It's a temporary implementation.
func (r Record) String() string {
	return fmt.Sprintf("took %s", r.Cost)
}
