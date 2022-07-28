package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/benchttp/engine/runner"
	"github.com/gorilla/websocket"
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
// Once the run is done, the state is updated. start sends message
// through the websocket connection, notifying the client.
func (r *run) start(ws *websocket.Conn) {
	r.flush()

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	r.runner = runner.New(r.sendProgess(ws))

	out, err := r.runner.Run(ctx, config())
	if err != nil {
		r.err = err
		_ = writeMessage(ws, fmt.Sprintf("done with error: %s", err))
		return
	}

	r.output = out
	_ = writeMessage(ws, "done without error")
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

// sendProgress sends the current runner.RecordingProgress through
// the websocket connection. As multiple goroutines may invoke sendProgess
// simultaneously as a callback via runner.onRecordingProgress, writing to
// the websocket connection is protected by a lock.
func (r *run) sendProgess(ws *websocket.Conn) func(runner.RecordingProgress) {
	return func(rp runner.RecordingProgress) {
		r.mu.Lock()
		defer r.mu.Unlock()

		m := fmt.Sprintf("%s: %d/%d %d", rp.Status(), rp.DoneCount, rp.MaxCount, rp.Percent())
		_ = writeMessage(ws, m)
	}
}

// sendOutput sends the output of the run through the websocket connection
// or a error message if there is no output available.
func (r *run) sendOutput(ws *websocket.Conn) {
	if r.output == nil {
		_ = writeMessage(ws, "not done yet")
		return
	}

	m := r.output.String()
	_ = writeMessage(ws, m)
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
