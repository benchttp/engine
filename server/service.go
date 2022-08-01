package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/benchttp/engine/internal/websocketio"
	"github.com/benchttp/engine/runner"
)

type service struct {
	mu     sync.RWMutex
	runner *runner.Runner
	cancel context.CancelFunc
}

// doRun calls runner.Runner.Run. The service state is overwritten.
// The return value of runner.Runner.Run is send to the client via
// w. The run progress is streamed through w.
func (s *service) doRun(w websocketio.Writer, cfg runner.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.runner = runner.New(s.sendRecordingProgess(w))

	out, err := s.runner.Run(ctx, cfg)
	if err != nil {
		_ = w.WriteJSON(errorMessage{Event: "done", Error: err.Error()})
		return
	}

	_ = w.WriteJSON(doneMessage{Event: "done", Data: *out})
}

// cancelRun cancels the run of the current runner.
// If the runner is nil, cancelRun is noop.
func (s *service) cancelRun() (ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runner == nil {
		return false
	}
	s.cancel()
	return true
}

// sendRecordingProgess returns a callback
// to send the current runner progress via w.
func (s *service) sendRecordingProgess(w websocketio.Writer) func(runner.RecordingProgress) {
	// The callback is invoked from a goroutine spawned by Recorder.Record.
	// Protect w from concurrent write with a lock.
	return func(rp runner.RecordingProgress) {
		s.mu.Lock()
		defer s.mu.Unlock()

		m := progressMessage{
			Event: "progress",
			Data:  fmt.Sprintf("%s: %d/%d %d", rp.Status(), rp.DoneCount, rp.MaxCount, rp.Percent()),
		}
		_ = w.WriteJSON(m)
	}
}

// flush clears the service state.
// Calling service.flush locks it for writing.
func (s *service) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runner = nil
	s.cancel = nil
}
