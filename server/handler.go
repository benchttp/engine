package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/internal/websocketio"
	"github.com/benchttp/engine/runner"
)

// Handler implements http.Handler.
// It serves a websocket server allowing
// remote manipulation of runner.Runner.
type Handler struct {
	mu      sync.Mutex
	Silent  bool
	Token   string
	service *service
}

func NewHandler(silent bool, token string) *Handler {
	return &Handler{
		Silent:  silent,
		Token:   token,
		service: &service{},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/run":
		h.handle(w, r)
	case "/stream":
		h.handleStream(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	upgrader := secureUpgrader(h.Token)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		ws.Close()
		// The client is gone, flush all the state.
		// TODO Handle reconnect?
		h.service.flush()
	}()

	log.Println("connected with client via websocket")

	rw := websocketio.NewReadWriter(ws, h.Silent)

	for {
		m := clientMessage{}
		if err := rw.ReadJSON(&m); err != nil {
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

			go h.service.doRun(rw, cfg)

		case "cancel":
			if ok := h.service.cancelRun(); !ok {
				rw.WriteJSON(errorMessage{Event: "error", Error: "not running"}) //nolint:errcheck
			}

		default:
			rw.WriteTextMessage(fmt.Sprintf("unknown procedure: %s", m.Action)) //nolint:errcheck
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

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		internalError(w, err)
		return
	}

	cfg, err := configparse.JSON(b)
	if err != nil {
		internalError(w, err)
		return
	}

	rep, err := runner.New(h.streamProgress(w)).Run(context.Background(), cfg)
	if err != nil {
		internalError(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(rep); err != nil {
		internalError(w, err)
		return
	}
}

func (h *Handler) streamProgress(w http.ResponseWriter) func(runner.RecordingProgress) {
	enc := json.NewEncoder(w)
	return func(progress runner.RecordingProgress) {
		if err := enc.Encode(progress); err != nil {
			internalError(w, err)
		}
		w.(http.Flusher).Flush()
	}
}

func internalError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
