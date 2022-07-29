package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/benchttp/engine/internal/configparse"
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
		inc := incomingMessage{}
		err := reader.ReadJSON(&inc)
		if err != nil {
			log.Println(err)
			break // Connection is dead.
		}

		// TODO Update package configparse for this purpose.
		p, err := json.Marshal(inc.Data)
		if err != nil {
			log.Println(err)
			break // Connection is dead.
		}
		cfg, err := configparse.JSON(p)
		if err != nil {
			log.Println(err)
			break // Connection is dead.
		}

		switch inc.Event {
		case "run":
			go run.start(writer, cfg)
			_ = writer.WriteJSON(outgoingMessage{Event: "running"})

		case "stop":
			ok := run.stop()
			if ok {
				_ = writer.WriteTextMessage("stopped")
			} else {
				_ = writer.WriteTextMessage("not running")
			}

		default:
			_ = writer.WriteTextMessage(fmt.Sprintf("unknown incoming event: %s", inc.Event))
		}
	}
}
