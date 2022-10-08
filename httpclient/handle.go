package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/benchttp/engine/configparse"
	"github.com/benchttp/engine/httpclient/response"
	"github.com/benchttp/engine/runner"
)

func handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		internalError(w, err)
		return
	}

	cfg, err := configparse.JSON(b)
	if err != nil {
		internalError(w, err)
		return
	}

	rep, err := runner.New(streamProgress(w)).Run(r.Context(), cfg)
	if err != nil {
		internalError(w, err)
		return
	}

	// hack: the write for the Report sometimes appears to be merged
	// with the last write for the Progress, leading to invalid JSON.
	// The issue is likely on the read side (front-end), but this is
	// the easiest fix for now.
	time.Sleep(10 * time.Millisecond)
	if err := response.Report(rep).EncodeJSON(w); err != nil {
		internalError(w, err)
		return
	}
}

func streamProgress(w http.ResponseWriter) func(runner.RecordingProgress) {
	return func(progress runner.RecordingProgress) {
		if err := response.Progress(progress).EncodeJSON(w); err != nil {
			internalError(w, err)
		}
		w.(http.Flusher).Flush()
	}
}

func internalError(w http.ResponseWriter, e error) {
	fmt.Fprint(os.Stderr, e)
	w.WriteHeader(http.StatusInternalServerError)

	if err := response.Error(e).EncodeJSON(w); err != nil {
		// Fallback to plain text encoding.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
