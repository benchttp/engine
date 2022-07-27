package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/benchttp/engine/runner"
	"github.com/gorilla/websocket"
)

type run struct {
	mu sync.RWMutex

	runner *runner.Runner
	output *runner.Report
	err    error
	cancel context.CancelFunc
}

func (r *run) run(ws *websocket.Conn) {
	r.flush()

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	r.runner = runner.New(
		func(rp runner.RecordingProgress) {
			// Protect from concurrent write to websocket connection.
			r.mu.Lock()
			defer r.mu.Unlock()
			m := fmt.Sprintf("%s: %d/%d %d", rp.Status(), rp.DoneCount, rp.MaxCount, rp.Percent())
			_ = writeMessage(ws, m)
		},
	)

	out, err := r.runner.Run(ctx, config())
	if err != nil {
		r.err = err
		_ = writeMessage(ws, fmt.Sprintf("done with error: %s", err))
		return
	}

	r.output = out
	_ = writeMessage(ws, "done without error")
}

func (r *run) stop() bool {
	defer r.flush()
	if r.runner == nil {
		return false
	}
	r.cancel()
	return true
}

func (r *run) pull(ws *websocket.Conn) {
	if r.output == nil {
		_ = writeMessage(ws, "not done yet")
		return
	}

	m := r.output.String()
	_ = writeMessage(ws, m)
}

func (r *run) flush() {
	r.runner = nil
	r.output = nil
	r.err = nil
	r.cancel = nil
}
