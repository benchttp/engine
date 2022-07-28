package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/benchttp/engine/server"
)

const port = "8080"

func main() {
	addr := ":" + port
	fmt.Println("http://localhost" + addr)

	handler := &server.Handler{
		Silent: false,
	}

	log.Fatal(http.ListenAndServe(addr, handler))
}
