package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/benchttp/engine/internal/cli/ansi"
	"github.com/benchttp/engine/runner"
)

// WriteRecordingProgress renders a fancy representation of s as a string
// and writes the result to w.
func WriteRecordingProgress(w io.Writer, s runner.RecordingProgress) (int, error) {
	return fmt.Fprint(w, renderProgress(s))
}

// renderProgress returns a string representation of runner.RecordingProgress
// for a fancy display in a CLI:
// 	RUNNING ◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎◼︎ 50% | 50/100 requests | 27s timeout
func renderProgress(s runner.RecordingProgress) string {
	var (
		countdown = s.Timeout - s.Elapsed
		reqmax    = strconv.Itoa(s.MaxCount)
		pctdone   = s.Percent()
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

// renderTimeline returns a colored representation of the progress as a string:
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
func renderStatus(status runner.RecordingStatus) string {
	color := statusStyle(status)
	return color(string(status))
}

func statusStyle(status runner.RecordingStatus) ansi.StyleFunc {
	switch status {
	case runner.StatusRunning:
		return ansi.Yellow
	case runner.StatusDone:
		return ansi.Green
	case runner.StatusCanceled:
		return ansi.Red
	case runner.StatusTimeout:
		return ansi.Cyan
	}
	return ansi.Grey // should not occur
}
