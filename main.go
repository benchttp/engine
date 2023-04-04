package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/benchttp/engine/cmd"
)


func main() {
	if err := cmd.Exec(); err != nil {
		fmt.Println(err)
		if errors.Is(err, cmd.ErrUsage) {
			flag.Usage()
		}
		os.Exit(1)
	}
}
