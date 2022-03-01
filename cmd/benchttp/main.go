package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var errUsage = errors.New("usage")

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		if errors.Is(err, errUsage) {
			flag.Usage()
		}
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("%w: no command specified", errUsage)
	}

	var cmd command
	args := os.Args[1:]

	switch sub := args[0]; sub {
	case "run":
		cmd = &cmdRun{flagset: flag.NewFlagSet("run", flag.ExitOnError)}
	case "auth":
		cmd = &cmdAuth{flagset: flag.NewFlagSet("auth", flag.ExitOnError)}
	default:
		return fmt.Errorf("%w: unknown command: %s", errUsage, sub)
	}

	return cmd.execute(args)
}

// command is the interface that all benchttp subcommands must implement.
type command interface {
	execute(args []string) error
}
