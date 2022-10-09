package response

import (
	"time"

	"github.com/benchttp/engine/runner"
)

func Progress(in runner.RecordingProgress) Response {
	return newResponse(progressResponse{
		Done:      in.Done,
		Error:     in.Error,
		DoneCount: in.DoneCount,
		MaxCount:  in.MaxCount,
		Timeout:   in.Timeout,
		Elapsed:   in.Elapsed,
	})
}

type progressResponse struct {
	Done      bool          `json:"done"`
	Error     error         `json:"error"`
	DoneCount int           `json:"doneCount"`
	MaxCount  int           `json:"maxCount"`
	Timeout   time.Duration `json:"timeout"`
	Elapsed   time.Duration `json:"elapsed"`
}
