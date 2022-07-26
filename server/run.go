package server

import (
	"io"
	"net/http"

	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/runner"
)

func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/run" {
		http.NotFound(w, r)
		return
	}

	// Allow single run at a time
	if s.isRequesterRunning() {
		http.Error(w, "already running", http.StatusConflict)
		return
	}
	defer s.flush()

	readBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cfg, err := configparse.JSON(readBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	report, err := s.doRun(silentConfig(cfg))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := report.WriteJSON(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func silentConfig(cfg runner.ConfigGlobal) runner.ConfigGlobal {
	cfg.Output = runner.ConfigOutput{
		Silent:   true,
		Template: "",
	}
	return cfg
}
