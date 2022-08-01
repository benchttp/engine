package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/internal/socketio"
	"github.com/benchttp/engine/runner"
	"github.com/gorilla/websocket"
)

// Handler has as single method, Handler.ServeHTTP.
// It serves a websocket server allowing remote
// manipulation of runner.Runner.
type Handler struct {
	Silent   bool
	Token    string
	service  *service
	upgrader websocket.Upgrader
}

func NewHandler(silent bool, token string) *Handler {
	return &Handler{
		Silent:  silent,
		Token:   token,
		service: &service{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return r.URL.Query().Get("access_token") == token
			},
		},
	}
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
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		ws.Close()
		// The client is gone, flush all the state.
		// TODO Handle reconnect?
		h.service.flush()
	}()

	log.Println("connected with client via websocket")

	reader := socketio.NewReader(ws, h.Silent)
	writer := socketio.NewWriter(ws, h.Silent)

	for {
		m := clientMessage{}
		err := reader.ReadJSON(&m)
		if err != nil {
			log.Println(err)
			break
		}

		switch m.Action {
		case "run":
			cfg, err := parseConfig(m.Data)
			if err != nil {
				log.Println(err)
				break
			}

			go h.service.doRun(writer, cfg)

		case "cancel":
			ok := h.service.cancelRun()
			if !ok {
				_ = writer.WriteJSON(errorMessage{Event: "error", Error: "not running"})
			}

		default:
			_ = writer.WriteTextMessage(fmt.Sprintf("unknown procedure: %s", m.Action))
		}
	}
}

// TODO Update package configparse for this purpose.

func parseConfig(data configparse.UnmarshaledConfig) (runner.Config, error) {
	p, err := json.Marshal(data)
	if err != nil {
		return runner.Config{}, err
	}

	cfg, err := configparse.JSON(p)
	if err != nil {
		log.Println(err)
		return runner.Config{}, err
	}

	return cfg, nil
}
