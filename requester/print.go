package requester

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/internal/cli/ansi"
)

// State represents the progression of a benchmark at a given time.
type State struct {
	Done                bool
	Error               error
	DoneCount, MaxCount int
	Timeout, Elapsed    time.Duration
}

// State returns the current State of the benchmark.
func (r *Requester) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return State{
		Done:      r.done,
		Error:     r.runErr,
		DoneCount: len(r.records),
		MaxCount:  r.config.Requests,
		Timeout:   r.config.GlobalTimeout,
		Elapsed:   time.Since(r.start),
	}
}

func (s State) JSON() ([]byte, error) {
	return json.Marshal(s)
}

// String returns a string representation of state for a fancy display
// in a CLI:
// 	RUNNING ◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎ 50% | 50/100 requests | 27s timeout
func (s State) String() string {
	var (
		countdown = s.Timeout - s.Elapsed
		reqmax    = strconv.Itoa(s.MaxCount)
		pctdone   = s.percentDone()
		timeline  = s.timeline(pctdone)
	)

	if reqmax == "-1" {
		reqmax = "∞"
	}
	if countdown < 0 {
		countdown = 0
	}

	return fmt.Sprintf(
		"%s%s %s %d%% | %d/%s requests | %.0fs timeout             \n",
		ansi.Erase(1),                 // replace previous line
		s.status(), timeline, pctdone, // progress
		s.DoneCount, reqmax, // requests
		countdown.Seconds(), // timeout
	)
}

var (
	tlBlock      = "◼︎"
	tlBlockGrey  = ansi.Grey(tlBlock)
	tlBlockGreen = ansi.Green(tlBlock)
	tlLen        = 10
)

// timeline returns a colored representation of the progress as a string:
// 	◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎
func (s State) timeline(pctdone int) string {
	tl := strings.Repeat(tlBlockGrey, tlLen)
	for i := 0; i < tlLen; i++ {
		if pctdone >= (tlLen * i) {
			tl = strings.Replace(tl, tlBlockGrey, tlBlockGreen, 1)
		}
	}
	return tl
}

// status returns a string representing the status, depending on whether
// the run is done or not and the value of the context error.
func (s State) status() string {
	if !s.Done {
		return ansi.Yellow("RUNNING")
	}
	switch s.Error {
	case nil:
		return ansi.Green("DONE")
	case context.Canceled:
		return ansi.Red("CANCELED")
	case context.DeadlineExceeded:
		return ansi.Cyan("TIMEOUT")
	}
	return "" // should not occur
}

// percentDone returns the progression of the run as a percentage.
// It is based on the ratio requests done / max requests if it's finite
// (not -1), else on the ratio elapsed time / global timeout.
func (s State) percentDone() int {
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
