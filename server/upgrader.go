package server

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func secureUpgrader(token string) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.URL.Query().Get("access_token") == token
		},
	}
}
