package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

func main() {
	// Try to bind to any available port.
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	port := l.Addr().(*net.TCPAddr).Port

	s := &http.Server{
		Addr:    ":" + fmt.Sprint(port),
		Handler: http.HandlerFunc(handleStream),
	}

	readyChan := make(chan struct{}, 1)
	closeChan := make(chan error, 1)

	go func() {
		readyChan <- struct{}{}
		closeChan <- (s.Serve(l))
	}()

	<-readyChan
	// Notify server is ready.
	fmt.Println("port:" + fmt.Sprint(port))
	fmt.Println("http://localhost:" + fmt.Sprint(port))

	log.Fatal(<-closeChan)
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
	log.Println(err.Error())

	w.WriteHeader(http.StatusInternalServerError)

	if err := json.NewEncoder(w).Encode(&struct {
		Error string
	}{Error: err.Error()}); err != nil {
		// Fallback to plain text encoding.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
