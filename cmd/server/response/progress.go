package response

import (
	"encoding/json"
	"io"
	"time"

	"github.com/benchttp/engine/runner"
)

type ProgressResponse struct {
	ID        int           `json:"id"`
	Done      bool          `json:"done"`
	Error     error         `json:"error"`
	DoneCount int           `json:"doneCount"`
	MaxCount  int           `json:"maxCount"`
	Timeout   time.Duration `json:"timeout"`
	Elapsed   time.Duration `json:"elapsed"`
}

func (resp ProgressResponse) EncodeJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

func Progress(in runner.RecordingProgress) ProgressResponse {
	return ProgressResponse{
		ID:        in.ID,
		Done:      in.Done,
		Error:     in.Error,
		DoneCount: in.DoneCount,
		MaxCount:  in.MaxCount,
		Timeout:   in.Timeout,
		Elapsed:   in.Elapsed,
	}
}
