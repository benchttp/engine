package main

import (
	"fmt"

	"github.com/benchttp/engine/server"
)

const port = "8080"

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {
	addr := ":" + port
	fmt.Println("http://localhost" + addr)
	return server.ListenAndServe(addr)
}
