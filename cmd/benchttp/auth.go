package main

import (
	"flag"
	"fmt"
)

// cmdAuth handles subcommand "benchttp auth [options]".
type cmdAuth struct {
	flagset *flag.FlagSet
}

// ensure cmdAuth implements command
var _ command = (*cmdAuth)(nil)

func (cmdAuth) execute(_ []string) error {
	fmt.Println("Benchttp authentication")
	return nil
}
