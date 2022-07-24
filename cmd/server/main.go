package main

import (
	"fmt"

	"github.com/benchttp/runner/server"
)

const port = "8080"

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {
	fmt.Println("http://localhost:" + port)
	return server.ListenAndServe(port)
}
