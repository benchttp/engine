package requester

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/runner/ansi"
)

// state represents the progression of a benchmark at a given time.
type state struct {
	done             bool
	err              error
	reqcur, reqmax   int
	timeout, elapsed time.Duration
}

// state returns the current state of the benchmark.
func (r *Requester) state() state {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return state{
		done:    r.done,
		err:     r.runErr,
		reqcur:  len(r.records),
		reqmax:  r.config.Requests,
		timeout: r.config.GlobalTimeout,
		elapsed: time.Since(r.start),
	}
}

// String returns a string representation of state for a fancy display
// in a CLI:
// 	RUNNING ◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎ 50% | 50/100 requests | 27s timeout
func (s state) String() string {
	var (
		countdown = s.timeout - s.elapsed
		reqmax    = strconv.Itoa(s.reqmax)
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
		s.reqcur, reqmax, // requests
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
func (s state) timeline(pctdone int) string {
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
func (s state) status() string {
	if !s.done {
		return ansi.Yellow("RUNNING")
	}
	switch s.err {
	case nil:
		return ansi.Green("DONE")
	case context.Canceled:
		return ansi.Cyan("CANCELED")
	case context.DeadlineExceeded:
		return ansi.Cyan("TIMEOUT")
	}
	return "" // should not occur
}

// percentDone returns the progression of the run as a percentage.
// It is based on the ratio requests done / max requests if it's finite
// (not -1), else on the ratio elapsed time / global timeout.
func (s state) percentDone() int {
	var cur, max int
	if s.reqmax == -1 {
		cur, max = int(s.elapsed), int(s.timeout)
	} else {
		cur, max = s.reqcur, s.reqmax
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
