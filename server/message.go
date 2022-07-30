package server

import (
	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

type clientMessage struct {
	Action string `json:"action"`
	// Data is non-empty if field Action is "run".
	Data configparse.UnmarshaledConfig `json:"data"`
}

type progressMessage struct {
	Event string `json:"state"`
	Data  string `json:"data"` // TODO runner.RecordingProgress
}

type doneMessage struct {
	Event string        `json:"state"`
	Data  runner.Report `json:"data"`
}

type errorMessage struct {
	Event string `json:"state"`
	Error string `json:"error"`
}
