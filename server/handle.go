package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/benchttp/engine/internal/socketio"
)

// Handler has as single method, Handler.ServeHTTP.
// It serves a websocket server.
type Handler struct {
	Silent bool
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/run":
		h.handle(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return // Connection is dead.
	}

	defer ws.Close()

	reader := socketio.NewReader(ws, h.Silent)
	writer := socketio.NewWriter(ws, h.Silent)

	log.Println("websocket connected with client")

	run := run{}
	defer run.flush()

	for {
		p, err := reader.ReadTextMessage()
		if err != nil {
			log.Println(err)
			break // Connection is dead.
		}
		m := string(p)

		switch m {
		case "run":
			go run.start(writer)
			_ = writer.WriteTextMessage("starting run")

		case "stop":
			ok := run.stop()
			if ok {
				_ = writer.WriteTextMessage("stopped")
			} else {
				_ = writer.WriteTextMessage("not running")
			}

		case "pull":
			run.sendOutput(writer)

		default:
			_ = writer.WriteTextMessage(fmt.Sprintf("unknown command: %s", m))
		}
	}
}
