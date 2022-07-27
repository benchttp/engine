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
	mu            sync.RWMutex
	currentRunner *runner.Runner
	stopRun       context.CancelFunc
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/run":
		s.handleRun(w, r)
	case "/state":
		s.handleState(w, r)
	case "/stop":
		s.handleStop(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) doRun(cfg runner.ConfigGlobal) (*runner.OutputReport, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s.setCurrentRequester(runner.New(nil))
	s.setStopRun(cancel)

	// Run benchmark
	return s.currentRunner.Run(ctx, silentConfig(cfg))
}

func (s *server) setCurrentRequester(r *runner.Runner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentRunner = r
}

func (s *server) setStopRun(cancelFunc context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopRun = cancelFunc
}

func (s *server) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentRunner = nil
	s.stopRun = nil
}

func (s *server) isRequesterRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentRunner != nil
}

func (s *server) requesterState() (state runner.RecorderProgress, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentRunner == nil {
		return runner.RecorderProgress{}, false
	}
	return s.currentRunner.Progress(), true
}

func (s *server) stopRequester() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentRunner == nil {
		return false
	}
	s.stopRun()
	return true
}

func silentConfig(cfg runner.ConfigGlobal) runner.ConfigGlobal {
	cfg.Output = runner.ConfigOutput{
		Silent:   true,
		Template: "",
	}
	return cfg
}
