package httpclient

import (
	"errors"
	"io"
	"log"
	"net/http"
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
		clientError(w, err)
		return
	}

	rep, err := runner.New(streamProgress(w)).Run(r.Context(), cfg)
	var invalidConfigError *runner.InvalidConfigError
	switch {
	case err == nil:
		// Pass through.
	case err == runner.ErrCanceled:
		clientError(w, err)
		return
	case errors.As(err, &invalidConfigError):
		clientError(w, invalidConfigError.Errors...)
		return
	default:
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

func clientError(w http.ResponseWriter, e ...error) {
	log.Printf("client error: %v\n", e)
	w.WriteHeader(http.StatusBadRequest)

	if err := response.ErrorClient(e).EncodeJSON(w); err != nil {
		// Fallback to plain text encoding.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func internalError(w http.ResponseWriter, e ...error) {
	log.Printf("server error: %v\n", e)
	w.WriteHeader(http.StatusInternalServerError)

	if err := response.ErrorServer(e).EncodeJSON(w); err != nil {
		// Fallback to plain text encoding.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
