package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/benchttp/engine/server"
)

const (
	port = "8080"
	// token is a dummy token used for development only.
	token = "6db67fafc4f5bf965a5a" //nolint:gosec
)

func main() {
	addr := ":" + port
	fmt.Println("http://localhost" + addr)

	handler := server.NewHandler(false, token)

	log.Fatal(http.ListenAndServe(addr, handler))
}
