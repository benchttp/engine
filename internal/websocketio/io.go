package websocketio

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Reader interface {
	ReadTextMessage() (string, error)
	ReadJSON(v interface{}) error
}

type Writer interface {
	WriteTextMessage(m string) error
	WriteJSON(v interface{}) error
}

type ReadWriter interface {
	Reader
	Writer
}

type readWriter struct {
	ws     *websocket.Conn
	silent bool
}

// NewReadWriter returns a concrete type ReadWriter
// reading from and writing to the websocket connection.
func NewReadWriter(ws *websocket.Conn, silent bool) ReadWriter {
	return &readWriter{ws, silent}
}

func (rw *readWriter) ReadTextMessage() (string, error) {
	messageType, p, err := rw.ws.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("cannot read message: %s", err)
	}

	if messageType != websocket.TextMessage {
		return "", fmt.Errorf("message type is not TextMessage")
	}

	m := string(p)

	if !rw.silent {
		log.Printf("<- %s", m)
	}

	return m, nil
}

func (rw *readWriter) ReadJSON(v interface{}) error {
	err := rw.ws.ReadJSON(v)
	if err != nil {
		return fmt.Errorf("cannot read message: %s", err)
	}

	if !rw.silent {
		log.Printf("<- %v", v)
	}

	return nil
}

func (rw *readWriter) WriteTextMessage(m string) error {
	err := rw.ws.WriteMessage(websocket.TextMessage, []byte(m))
	if err != nil {
		return fmt.Errorf("cannot write message: %s", err)
	}

	if !rw.silent {
		log.Printf("-> %s", m)
	}

	return nil
}

func (rw *readWriter) WriteJSON(v interface{}) error {
	err := rw.ws.WriteJSON(v)
	if err != nil {
		return fmt.Errorf("cannot write message: %s", err)
	}

	if !rw.silent {
		log.Printf("-> %v", v)
	}

	return nil
}
