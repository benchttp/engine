package server

import (
	"net/http"

	"github.com/gorilla/websocket"
)

const token = "6db67fafc4f5bf965a5a" //nolint:gosec

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.URL.Query().Get("access_token") == token
	},
}
