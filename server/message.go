package server

import (
	"github.com/benchttp/engine/internal/configparse"
)

type incomingMessage struct {
	Event string                        `json:"event"`
	Data  configparse.UnmarshaledConfig `json:"data"`
}

type outgoingMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
