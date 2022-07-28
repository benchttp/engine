package socketio

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Reader interface {
	ReadTextMessage() (string, error)
	ReadJSON() error
}

type reader struct {
	ws     *websocket.Conn
	silent bool
}

// NewReader returns a concrete type Reader that will read from
// the websocket connection.
func NewReader(ws *websocket.Conn, slient bool) Reader {
	return &reader{ws, slient}
}

func (r *reader) ReadTextMessage() (string, error) {
	messageType, p, err := r.ws.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("cannot read message: %s", err)
	}

	if messageType != websocket.TextMessage {
		return "", fmt.Errorf("message type is not TextMessage")
	}

	m := string(p)

	if !r.silent {
		log.Printf("<- %s", m)
	}

	return m, nil
}

func (r *reader) ReadJSON() error {
	return fmt.Errorf("not implemented")
}
