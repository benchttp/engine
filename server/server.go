package server

import "net/http"

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, runHandler{})
}
