package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/benchttp/engine/config"
	"github.com/benchttp/engine/output"
	"github.com/benchttp/engine/requester"
)

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, &server{})
}

type server struct {
	mu               sync.RWMutex
	currentRequester *requester.Requester
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/run":
		s.handleRun(w, r)
	case "/state":
		s.handleState(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) doRun(cfg config.Global) (output.Report, error) {
	// Retrieve HTTP request generated by the config
	httpRequest, err := cfg.Request.Value()
	if err != nil {
		return output.Report{}, err
	}

	s.setCurrentRequester(requester.New(requesterConfig(cfg)))

	// Run benchmark
	bk, err := s.currentRequester.Run(context.Background(), httpRequest)
	if err != nil {
		return output.Report{}, err
	}

	return *output.New(bk, cfg, ""), nil
}

func (s *server) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentRequester = nil
}

func (s *server) isRequesterRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentRequester != nil
}

func (s *server) setCurrentRequester(r *requester.Requester) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentRequester = r
}

func (s *server) requesterState() (state requester.State, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isRequesterRunning() {
		return requester.State{}, false
	}
	return s.currentRequester.State(), true
}

// requesterConfig returns a requester.Config generated from cfg.
func requesterConfig(cfg config.Global) requester.Config {
	return requester.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		Silent:         cfg.Output.Silent,
	}
}
