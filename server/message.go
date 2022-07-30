package server

import (
	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

type clientMessage struct {
	Action string `json:"action"`
	// Data is non-empty if MessageProcedure.Action is "start".
	Data configparse.UnmarshaledConfig `json:"data"`
}

type progressMessage struct {
	Event string `json:"state"`
	// Data  runner.RecordingProgress `json:"data"`
	Data string `json:"data"`
}

type doneMessage struct {
	Event string        `json:"state"`
	Data  runner.Report `json:"data"`
}

type errorMessage struct {
	Event string `json:"state"`
	Error string `json:"error"`
}
