package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

const readySignal = "READY"

var (
	// stdout is used to send messages to the parent process.
	// Use stdout to send info and debug level messages.
	stdout = log.New(os.Stdout, "", 0)
	// stderr is used to send messages to the parent process.
	// Use stderr to send error level messages or to notify
	// any action which are followed by os.Exit.
	stderr = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	useAutoPort := flag.Bool("auto-port", true, "automatically find and use an available port")
	flag.Parse()

	var p string
	if *useAutoPort {
		p = "0"
	} else {
		err := godotenv.Load("./.env.development")
		if err != nil {
			stderr.Fatalf("could not load .env file: %s", err.Error())
		}
		p = os.Getenv("SERVER_PORT")
	}

	addr, err := net.ResolveTCPAddr("tcp", "localhost:"+p)
	if err != nil {
		stderr.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		stderr.Fatal(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	s := &http.Server{
		Addr:    ":" + fmt.Sprint(port),
		Handler: http.HandlerFunc(handleStream),
		// No timeout because the end user runs both the client and
		// the server on their machine, which makes it irrelevant.
		ReadHeaderTimeout: 0,
	}

	// Notify server is ready.
	stdout.Println(readySignalString(port))

	stderr.Fatal(s.Serve(listener))
}

func readySignalString(port int) string {
	return fmt.Sprintf("%s http://localhost:%d", readySignal, port)
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
	if err := toReportResponse(rep).EncodeJSON(w); err != nil {
		internalError(w, err)
		return
	}
}

func streamProgress(w http.ResponseWriter) func(runner.RecordingProgress) {
	return func(progress runner.RecordingProgress) {
		if err := toProgressResponse(progress).EncodeJSON(w); err != nil {
			internalError(w, err)
		}
		w.(http.Flusher).Flush()
	}
}

func internalError(w http.ResponseWriter, e error) {
	stderr.Println(e)

	w.WriteHeader(http.StatusInternalServerError)

	if err := toErrorResponse(e).EncodeJSON(w); err != nil {
		// Fallback to plain text encoding.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
