package requester

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/runner/ansi"
)

type state struct {
	done             bool
	err              error
	reqcur, reqmax   int
	timeout, elapsed time.Duration
}

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
		ansi.Erase(1), s.status(), timeline, pctdone, // progress
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

func (s state) timeline(pctdone int) string {
	tl := strings.Repeat(tlBlockGrey, tlLen)
	for i := 0; i < tlLen; i++ {
		if pctdone >= (tlLen * i) {
			tl = strings.Replace(tl, tlBlockGrey, tlBlockGreen, 1)
		}
	}
	return tl
}

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

func (s state) percentDone() int {
	var cur, max int
	if s.reqmax == -1 {
		cur, max = int(s.elapsed), int(s.timeout)
	} else {
		cur, max = s.reqcur, s.reqmax
	}
	return capInt((100*cur)/max, 100)
}

func capInt(n, max int) int {
	if n > max {
		return max
	}
	return n
}
