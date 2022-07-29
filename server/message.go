package server

import (
	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

type messageProcedure struct {
	Procedure string `json:"procedure"`
	// Data is non-empty if MessageProcedure.Procedure is "start".
	Data configparse.UnmarshaledConfig `json:"data"`
}

type messageProgress struct {
	Event string `json:"event"`
	// Data  runner.RecordingProgress `json:"data"`
	Data string `json:"data"`
}

type messageDone struct {
	Event string        `json:"event"`
	Data  runner.Report `json:"data"`
}

type messageError struct {
	Event string `json:"event"`
	Error error  `json:"error"`
}
