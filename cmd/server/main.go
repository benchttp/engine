package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/benchttp/engine/server"
)

const (
	defaultPort = "8080"
	// token is a dummy token used for development only.
	token = "6db67fafc4f5bf965a5a" //nolint:gosec
)

var port string

func init() {
	flag.StringVar(&port, "port", defaultPort, "port to listen on")
	flag.Parse()
}

func main() {
	addr := ":" + port
	fmt.Println("http://localhost" + addr)

	handler := server.NewHandler(false, token)

	log.Fatal(http.ListenAndServe(addr, handler))
}
