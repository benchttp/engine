package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/benchttp/engine/internal/socketio"
	"github.com/benchttp/engine/runner"
)

type service struct {
	mu sync.RWMutex

	runner *runner.Runner
	cancel context.CancelFunc
}

// doRun calls to runner.Run. Any previous state is immediately flushed.
// Once the doRun is done, the state is updated. doRun uses w to notify
// that the doRun has started and the done status upon completion.
func (s *service) doRun(w socketio.Writer, cfg runner.Config) {
	s.flush()

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

// cancelRun stops the run if it is running. The state is always flushed.
func (s *service) cancelRun() (ok bool) {
	ok = s.runner != nil
	s.flush()
	return
}

// sendRecordingProgess sends the current runner.RecordingProgress via w.
// As multiple goroutines may invoke run.sendRecordingProgess concurrently
// as a callback from runner.onRecordingProgress, writing to w
// is protected by a lock.
func (s *service) sendRecordingProgess(w socketio.Writer) func(runner.RecordingProgress) {
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

// flush clears the state. Calling run.flush locks the run for writing.
func (s *service) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}
	s.runner = nil
	s.cancel = nil
}
