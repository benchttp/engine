package server

import (
	"encoding/json"
	"errors"
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
		return
	}

	defer ws.Close()

	log.Println("websocket connected with client")

	reader := socketio.NewReader(ws, h.Silent)
	writer := socketio.NewWriter(ws, h.Silent)

	run := run{}
	defer run.flush()

	for {
		inc := messageProcedure{}
		err := reader.ReadJSON(&inc)
		if err != nil {
			log.Println(err)
			break
		}

		switch inc.Procedure {
		case "run":
			// TODO Update package configparse for this purpose.
			p, err := json.Marshal(inc.Data)
			if err != nil {
				log.Println(err)
				break
			}
			cfg, err := configparse.JSON(p)
			if err != nil {
				log.Println(err)
				break
			}

			go run.start(writer, cfg)

		case "stop":
			ok := run.stop()
			if !ok {
				_ = writer.WriteJSON(messageError{Event: "error", Error: errors.New("not running")})
			}

		default:
			_ = writer.WriteTextMessage(fmt.Sprintf("unknown procedure: %s", inc.Procedure))
		}
	}
}
