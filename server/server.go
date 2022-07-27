package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/benchttp/engine/runner"
)

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, &server{})
}

type server struct {
	mu      sync.RWMutex
	runner  *runner.Runner
	stopRun context.CancelFunc
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/run":
		s.handleRun(w, r)
	case "/progress":
		s.handleProgress(w, r)
	case "/stop":
		s.handleStop(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) doRun(cfg runner.Config) (*runner.Report, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s.setRunner(runner.New(nil))
	s.setStopRun(cancel)

	// Run benchmark
	return s.runner.Run(ctx, silentConfig(cfg))
}

func (s *server) setRunner(r *runner.Runner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runner = r
}

func (s *server) setStopRun(cancelFunc context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopRun = cancelFunc
}

func (s *server) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runner = nil
	s.stopRun = nil
}

func (s *server) isRequesterRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runner != nil
}

func (s *server) recordingProgress() (progress runner.RecordingProgress, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.runner == nil {
		return runner.RecordingProgress{}, false
	}
	return s.runner.Progress(), true
}

func (s *server) stopRequester() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runner == nil {
		return false
	}
	s.stopRun()
	return true
}

func silentConfig(cfg runner.Config) runner.Config {
	cfg.Output = runner.OutputConfig{
		Silent:   true,
		Template: "",
	}
	return cfg
}
