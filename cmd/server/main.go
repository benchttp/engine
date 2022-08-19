package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

const (
	port = "8080"
)

func main() {
	addr := ":" + port
	fmt.Println("http://localhost" + addr)

	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handleStream)))
}

func handleStream(w http.ResponseWriter, r *http.Request) {
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
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		internalError(w, err)
		return
	}
}

func streamProgress(w http.ResponseWriter) func(runner.RecordingProgress) {
	enc := json.NewEncoder(w)
	return func(progress runner.RecordingProgress) {
		if err := enc.Encode(progress); err != nil {
			internalError(w, err)
		}
		w.(http.Flusher).Flush()
	}
}

func internalError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
