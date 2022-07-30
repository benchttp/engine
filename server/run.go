package server

import (
	"io"
	"net/http"

	"github.com/benchttp/engine/internal/configfile"
)

func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	// Allow single run at a time
	if s.isRequesterRunning() {
		http.Error(w, "already running", http.StatusConflict)
		return
	}
	defer s.flush()

	// Read input config
	readBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse json config
	cfg, err := configfile.JSON(readBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start run
	out, err := s.doRun(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with run output
	if _, err := out.WriteJSON(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
