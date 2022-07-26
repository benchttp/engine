package recorder

import (
	"context"
	"encoding/json"
	"time"
)

type Status string

const (
	StatusRunning  Status = "RUNNING"
	StatusCanceled Status = "CANCELED"
	StatusTimeout  Status = "TIMEOUT"
	StatusDone     Status = "DONE"
)

// Progress represents the progression of a recording at a given time.
type Progress struct {
	Done                bool
	Error               error
	DoneCount, MaxCount int
	Timeout, Elapsed    time.Duration
}

// Progress returns the current Progress of the recording.
func (r *Recorder) Progress() Progress {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return Progress{
		Done:      r.done,
		Error:     r.runErr,
		DoneCount: len(r.records),
		MaxCount:  r.config.Requests,
		Timeout:   r.config.GlobalTimeout,
		Elapsed:   time.Since(r.start),
	}
}

func (s Progress) JSON() ([]byte, error) {
	return json.Marshal(s)
}

// status returns a string representing the status, depending on whether
// the run is done or not and the value of the context error.
func (s Progress) Status() Status {
	if !s.Done {
		return StatusRunning
	}
	switch s.Error {
	case nil:
		return StatusDone
	case context.Canceled:
		return StatusCanceled
	case context.DeadlineExceeded:
		return StatusTimeout
	}
	return "" // should not occur
}

// percentDone returns the progression of the run as a percentage.
// It is based on the ratio requests done / max requests if it's finite
// (not -1), else on the ratio elapsed time / global timeout.
func (s Progress) Percent() int {
	var cur, max int
	if s.MaxCount == -1 {
		cur, max = int(s.Elapsed), int(s.Timeout)
	} else {
		cur, max = s.DoneCount, s.MaxCount
	}
	return capInt((100*cur)/max, 100)
}

// capInt returns n if n <= max, max otherwise.
func capInt(n, max int) int {
	if n > max {
		return max
	}
	return n
}
