package server

import "net/http"

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, ok := s.requesterState()
	if !ok {
		http.Error(w, "not running", http.StatusConflict)
		return
	}

	jsonState, err := state.JSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonState)
}
