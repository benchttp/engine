package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/benchttp/engine/internal/socketio"
	"github.com/benchttp/engine/runner"
)

// run is a stateful representation of the current run
// performed by the server.
type run struct {
	mu sync.RWMutex

	runner *runner.Runner
	output *runner.Report
	err    error
	cancel context.CancelFunc
}

// start starts the run. Any previous state is immediately flushed.
// Once the run is done, the state is updated. start uses w to notify
// that the run has started and the done status upon completion.
func (r *run) start(w socketio.Writer) {
	r.flush()

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	r.runner = runner.New(r.sendProgess(w))

	out, err := r.runner.Run(ctx, config())
	if err != nil {
		r.err = err
		_ = w.WriteTextMessage(fmt.Sprintf("done with error: %s", err))
		return
	}

	r.output = out
	_ = w.WriteTextMessage("done without error")
}

// stop stops the run if it is running. The state is always flushed.
func (r *run) stop() bool {
	defer r.flush()
	if r.runner == nil {
		return false
	}
	r.cancel()
	return true
}

// sendProgress sends the current runner.RecordingProgress via w.
// As multiple goroutines may invoke sendProgess simultaneously
// as a callback from runner.onRecordingProgress, writing to w
// is protected by a lock.
func (r *run) sendProgess(w socketio.Writer) func(runner.RecordingProgress) {
	return func(rp runner.RecordingProgress) {
		r.mu.Lock()
		defer r.mu.Unlock()

		m := fmt.Sprintf("%s: %d/%d %d", rp.Status(), rp.DoneCount, rp.MaxCount, rp.Percent())
		_ = w.WriteTextMessage(m)
	}
}

// sendOutput sends the output of the run via w or an error message
// if there is no output available.
func (r *run) sendOutput(w socketio.Writer) {
	if r.output == nil {
		_ = w.WriteTextMessage("not done yet")
		return
	}

	_ = w.WriteTextMessage(r.output.String())
}

// flush clears the state.
func (r *run) flush() {
	if r.cancel != nil {
		r.cancel()
	}
	r.runner = nil
	r.output = nil
	r.err = nil
	r.cancel = nil
}
