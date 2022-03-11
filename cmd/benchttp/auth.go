package main

import (
	"flag"
	"fmt"

	"github.com/benchttp/runner/ansi"
	"github.com/benchttp/runner/internal/auth"
)

// tokenURL is the URL to the webapp where the user can get a token.
const tokenURL = "https://www.benchttp.app/login" // nolint:gosec // no creds

// cmdAuth handles subcommand "benchttp auth [options]".
type cmdAuth struct {
	flagset *flag.FlagSet
}

func (cmd cmdAuth) execute(args []string) error {
	if len(args) != 2 {
		cmd.flagset.Usage()
		return errUsage
	}

	switch sub := args[1]; sub {
	case "login":
		return cmd.login()
	case "logout":
		return cmd.logout()
	default:
		return fmt.Errorf("%w: unknown subcommand: %s", errUsage, sub)
	}
}

func (cmd cmdAuth) login() error {
	token, err := promptf("Visit %s and paste the token:\n", tokenURL)
	if err != nil {
		return err
	}

	if err := auth.SaveToken(token); err != nil {
		return err
	}

	fmt.Printf("%sSuccessfully logged in.\n", ansi.Erase(1))
	return nil
}

func (cmd cmdAuth) logout() error {
	if err := auth.DeleteToken(); err != nil {
		return err
	}

	fmt.Println("Successfully logged out.")
	return nil
}
