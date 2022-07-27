package server

import "net/http"

func (s *server) handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	progress, ok := s.recordingProgress()
	if !ok {
		http.Error(w, "not running", http.StatusConflict)
		return
	}

	jsonProgress, err := progress.JSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonProgress)
}
