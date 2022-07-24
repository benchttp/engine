package server

import (
	"io"
	"net/http"

	"github.com/benchttp/engine/config"
	"github.com/benchttp/engine/internal/configparse"
)

func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/run" {
		http.NotFound(w, r)
		return
	}

	// Allow single run at a tinme
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

	jsonReport, err := report.JSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonReport)
}

func silentConfig(cfg config.Global) config.Global {
	cfg.Output = config.Output{
		Silent:   true,
		Out:      []config.OutputStrategy{},
		Template: "",
	}
	return cfg
}
