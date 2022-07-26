package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/benchttp/engine/internal/cli/ansi"
	"github.com/benchttp/engine/requester"
)

// WriteRequesterState renderes a fancy representation of s as a string
// and writes the result to w.
func WriteRequesterState(w io.Writer, s requester.State) (int, error) {
	return fmt.Fprint(w, renderState(s))
}

// renderState returns a string representation of requester.State
// for a fancy display in a CLI:
// 	RUNNING ◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎ 50% | 50/100 requests | 27s timeout
func renderState(s requester.State) string {
	var (
		countdown = s.Timeout - s.Elapsed
		reqmax    = strconv.Itoa(s.MaxCount)
		pctdone   = s.PercentDone()
		timeline  = renderTimeline(pctdone)
	)

	if reqmax == "-1" {
		reqmax = "∞"
	}
	if countdown < 0 {
		countdown = 0
	}

	return fmt.Sprintf(
		"%s%s %s %d%% | %d/%s requests | %.0fs timeout             \n",
		ansi.Erase(1),                               // replace previous line
		renderStatus(s.Status()), timeline, pctdone, // progress
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
func renderTimeline(pctdone int) string {
	tl := strings.Repeat(tlBlockGrey, tlLen)
	for i := 0; i < tlLen; i++ {
		if pctdone >= (tlLen * i) {
			tl = strings.Replace(tl, tlBlockGrey, tlBlockGreen, 1)
		}
	}
	return tl
}

// renderStatus returns a string representing the status,
// depending on whether the run is done or not and the value
// of its context error.
func renderStatus(status requester.Status) string {
	color := statusStyle(status)
	return color(string(status))
}

func statusStyle(status requester.Status) ansi.StyleFunc {
	switch status {
	case requester.StatusRunning:
		return ansi.Yellow
	case requester.StatusDone:
		return ansi.Green
	case requester.StatusCanceled:
		return ansi.Red
	case requester.StatusTimeout:
		return ansi.Cyan
	}
	return ansi.Grey // should not occur
}