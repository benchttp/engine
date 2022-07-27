package server

import (
	"fmt"
	"log"
	"net/http"
)

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/run":
		handle(w, r)
	default:
		http.NotFound(w, r)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return // Connection is dead.
	}

	defer ws.Close()

	log.Println("websocket connected with client")

	run := run{}

	for {
		m, err := readMessage(ws)
		if err != nil {
			log.Println(err)
			break // Connection is dead.
		}

		switch m {
		case "run":
			go run.run(ws)
			_ = writeMessage(ws, "starting run")

		case "stop":
			ok := run.stop()
			if ok {
				_ = writeMessage(ws, "stopped")
			} else {
				_ = writeMessage(ws, "not running")
			}

		case "pull":
			run.pull(ws)

		default:
			_ = writeMessage(ws, fmt.Sprintf("unknown command: %s", m))
		}
	}

	// Clean up
	if run.cancel != nil {
		run.cancel()
	}
	run.flush()
}
