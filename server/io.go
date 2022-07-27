package server

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func readMessage(ws *websocket.Conn) (string, error) {
	messageType, m, err := ws.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("cannot read message: %s", err)
	}

	if messageType != websocket.TextMessage {
		return "", fmt.Errorf("message type is not TextMessage")
	}

	ms := string(m)

	log.Printf("<- %s", ms)

	return ms, nil
}

func writeMessage(ws *websocket.Conn, m string) error {
	mb := []byte(m)

	err := ws.WriteMessage(websocket.TextMessage, mb)
	if err != nil {
		log.Println("cannot write message:", err)
		return err
	}

	log.Printf("-> %s", m)

	return nil
}
