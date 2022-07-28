package socketio

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Writer interface {
	WriteTextMessage(m string) error
	WriteJSON(v interface{}) error
}

type writer struct {
	ws     *websocket.Conn
	silent bool
}

// NewWriter returns a concrete type Writer that will write to
// the websocket connection.
func NewWriter(ws *websocket.Conn, silent bool) Writer {
	return &writer{ws, silent}
}

func (w *writer) WriteTextMessage(m string) error {
	err := w.ws.WriteMessage(websocket.TextMessage, []byte(m))
	if err != nil {
		return fmt.Errorf("cannot write message: %s", err)
	}

	if !w.silent {
		log.Printf("-> %s", m)
	}

	return nil
}

func (w *writer) WriteJSON(v interface{}) error {
	err := w.ws.WriteJSON(v)
	if err != nil {
		return fmt.Errorf("cannot write message: %s", err)
	}

	if !w.silent {
		log.Printf("-> %v", v)
	}

	return nil
}
