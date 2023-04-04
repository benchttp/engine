package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// errUsage reports an incorrect usage of the benchttp command.
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
	commandName, options, err := shiftArgs(os.Args[1:])
	if err != nil {
		return err
	}

	cmd, err := commandOf(commandName)
	if err != nil {
		return err
	}

	return cmd.execute(options)
}

func shiftArgs(args []string) (commandName string, nextArgs []string, err error) {
	if len(args) < 1 {
		return "", []string{}, fmt.Errorf("%w: no command specified", errUsage)
	}
	return args[0], args[1:], nil
}

// command is the interface that all benchttp subcommands must implement.
type command interface {
	execute(args []string) error
}

func commandOf(name string) (command, error) {
	switch name {
	case "run":
		return &cmdRun{flagset: flag.NewFlagSet("run", flag.ExitOnError)}, nil
	case "version":
		return &cmdVersion{}, nil
	default:
		return nil, fmt.Errorf("%w: unknown command: %s", errUsage, name)
	}
}
