package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/benchttp/engine/internal/cli"
	"github.com/benchttp/engine/internal/cli/configflags"
	"github.com/benchttp/engine/internal/configparse"
	"github.com/benchttp/engine/internal/signals"
	"github.com/benchttp/engine/runner"
)

// cmdRun handles subcommand "benchttp run [options]".
type cmdRun struct {
	flagset *flag.FlagSet

	// configFile is the parsed value for flag -configFile
	configFile string

	// config is the runner config resulting from parsing CLI flags.
	config runner.Config
}

// init initializes cmdRun with default values.
func (cmd *cmdRun) init() {
	cmd.config = runner.DefaultConfig()
	cmd.configFile = configparse.Find([]string{
		"./.benchttp.yml",
		"./.benchttp.yaml",
		"./.benchttp.json",
	})
}

// execute runs the benchttp runner: it parses CLI flags, loads config
// from config file and parsed flags, then runs the benchmark and outputs
// it according to the config.
func (cmd *cmdRun) execute(args []string) error {
	cmd.init()

	// Set CLI config from flags and retrieve fields that were set
	fieldsSet := cmd.parseArgs(args)

	// Generate merged config (defaults < config file < CLI flags)
	cfg, err := cmd.makeConfig(fieldsSet)
	if err != nil {
		return err
	}

	// Prepare graceful shutdown in case of os.Interrupt (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	go signals.ListenOSInterrupt(cancel)

	// Run the benchmark
	out, err := runner.
		New(onRecordingProgress(cfg.Output.Silent)).
		Run(ctx, cfg)
	if err != nil {
		return err
	}

	// Write results to stdout
	if _, err := out.Write(os.Stdout); err != nil {
		return err
	}

	return nil
}

// parseArgs parses input args as config fields and returns
// a slice of fields that were set by the user.
func (cmd *cmdRun) parseArgs(args []string) []string {
	// first arg is subcommand "run"
	// skip parsing if no flags are provided
	if len(args) <= 1 {
		return []string{}
	}

	// config file path
	cmd.flagset.StringVar(&cmd.configFile,
		"configFile",
		cmd.configFile,
		"Config file path",
	)

	// attach config options flags to the flagset
	// and bind their value to the config struct
	configflags.Bind(cmd.flagset, &cmd.config)

	cmd.flagset.Parse(args[1:]) //nolint:errcheck // never occurs due to flag.ExitOnError

	return configflags.Which(cmd.flagset)
}

// makeConfig returns a runner.ConfigGlobal initialized with config file
// options if found, overridden with CLI options listed in fields
// slice param.
func (cmd *cmdRun) makeConfig(fields []string) (cfg runner.Config, err error) {
	cliConfig := cmd.config.WithFields(fields...)

	// configFile not set and default ones not found:
	// skip the merge and return the cli config
	if cmd.configFile == "" {
		return cliConfig, cliConfig.Validate()
	}

	fileConfig, err := configparse.Parse(cmd.configFile)
	if err != nil && !errors.Is(err, configparse.ErrFileNotFound) {
		// config file is not mandatory: discard ErrFileNotFound.
		// other errors are critical
		return
	}

	mergedConfig := fileConfig.Override(cliConfig)

	return mergedConfig, mergedConfig.Validate()
}

func onRecordingProgress(silent bool) func(runner.RecordingProgress) {
	if silent {
		return func(runner.RecordingProgress) {}
	}

	// hack: write a blank line as cli.WriteRecordingProgress always
	// erases the previous line
	fmt.Println()

	return func(progress runner.RecordingProgress) {
		cli.WriteRecordingProgress(os.Stdout, progress) //nolint: errcheck
	}
}
