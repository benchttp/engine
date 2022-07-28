package server

type message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
