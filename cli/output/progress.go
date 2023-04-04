package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/benchttp/engine/benchttp"

	"github.com/benchttp/engine/cli/output/ansi"
)

// Progress renders a fancy representation of a runner.RecordingProgress
// and writes the result to w.
func Progress(w io.Writer, p benchttp.RecordingProgress) (int, error) {
	return fmt.Fprint(w, progressString(p))
}

// progressString returns a string representation of a runner.RecordingProgress
// for a fancy display in a CLI:
//
//	RUNNING ◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎ 50% | 50/100 requests | 27s timeout
func progressString(p benchttp.RecordingProgress) string {
	var (
		countdown = p.Timeout - p.Elapsed
		reqmax    = strconv.Itoa(p.MaxCount)
		pctdone   = p.Percent()
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
		renderStatus(p.Status()), timeline, pctdone, // progress
		p.DoneCount, reqmax, // requests
		countdown.Seconds(), // timeout
	)
}

var (
	tlBlock      = "◼︎"
	tlBlockGrey  = ansi.Grey(tlBlock)
	tlBlockGreen = ansi.Green(tlBlock)
	tlLen        = 10
)

// renderTimeline returns a colored representation of the progress as a string:
//
//	◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎
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
func renderStatus(status benchttp.RecordingStatus) string {
	styled := statusStyle(status)
	return styled(string(status))
}

func statusStyle(status benchttp.RecordingStatus) ansi.StyleFunc {
	switch status {
	case benchttp.StatusRunning:
		return ansi.Yellow
	case benchttp.StatusDone:
		return ansi.Green
	case benchttp.StatusCanceled:
		return ansi.Red
	case benchttp.StatusTimeout:
		return ansi.Cyan
	}
	return ansi.Grey // should not occur
}
