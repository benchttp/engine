package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// ErrUsage reports an incorrect usage of the benchttp-run command.
var ErrUsage = errors.New("usage")

func Exec() error {
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
		return "", []string{}, fmt.Errorf("%w: no command specified", ErrUsage)
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
		return nil, fmt.Errorf("%w: unknown command: %s", ErrUsage, name)
	}
}
