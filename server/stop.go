package server

import "net/http"

func (s *server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer s.flush()

	if ok := s.stopRequester(); !ok {
		http.Error(w, "not running", http.StatusConflict)
		return
	}
}
