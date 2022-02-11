package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/benchttp/runner/config"
	configfile "github.com/benchttp/runner/config/file"
	"github.com/benchttp/runner/requester"
)

const (
	// API server endpoint. May live is some config file later.
	reportURL = "http://localhost:9998/report"
)

var (
	configFile    string
	uri           string
	header        = http.Header{}
	concurrency   int           // Number of connections to run concurrently
	requests      int           // Number of requests to run, use duration as exit condition if omitted.
	timeout       time.Duration // Timeout for each HTTP request
	interval      time.Duration // Minimum duration between two groups of requests
	globalTimeout time.Duration // Duration of test
)

var defaultConfigFiles = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

func parseArgs() {
	// config file path
	flag.StringVar(&configFile, "configFile", configfile.Find(defaultConfigFiles), "Config file path")

	// request url
	flag.StringVar(&uri, config.FieldURL, "", "Target URL to request")
	// request header
	flag.Var(headerValue{header: &header}, config.FieldHeader, "HTTP request header")
	// request timeout
	flag.DurationVar(&timeout, config.FieldTimeout, 0, "Timeout for each HTTP request")
	// concurrency
	flag.IntVar(&concurrency, config.FieldConcurrency, 0, "Number of connections to run concurrently")
	// requests number
	flag.IntVar(&requests, config.FieldRequests, 0, "Number of requests to run, use duration as exit condition if omitted")
	// non-conurrent requests interval
	flag.DurationVar(&interval, "interval", 0, "Minimum duration between two non concurrent requests")
	// global timeout
	flag.DurationVar(&globalTimeout, config.FieldGlobalTimeout, 0, "Max duration of test")

	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	parseArgs()

	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	if err := requester.New(cfg).RunAndSendReport(reportURL); err != nil {
		return err
	}

	return nil
}

// parseConfig returns a config.Config initialized with config file
// options if found, overridden with CLI options.
func parseConfig() (cfg config.Config, err error) {
	fileCfg, err := configfile.Parse(configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory, other errors are critical
		return
	}

	cliCfg := config.Config{
		Request: config.Request{
			Header:  header,
			Timeout: timeout,
		},
		RunnerOptions: config.RunnerOptions{
			Requests:      requests,
			Concurrency:   concurrency,
			Interval:      interval,
			GlobalTimeout: globalTimeout,
		},
	}.WithURL(uri)

	mergedConfig := fileCfg.Override(cliCfg, configFlags()...)

	return mergedConfig, mergedConfig.Validate()
}

// configFlags returns a slice of all config fields set via the CLI.
func configFlags() []string {
	var fields []string
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if name := f.Name; config.IsField(name) {
			fields = append(fields, name)
		}
	})
	return fields
}

// headerValue implements flag.Value
type headerValue struct {
	header *http.Header
}

// String returns a string representation of the referenced header.
func (v headerValue) String() string {
	return fmt.Sprint(v.header)
}

// Set reads input string in format "key:value" and appends value
// to the key's values of the referenced header.
func (v headerValue) Set(in string) error {
	keyval := strings.Split(in, ":")
	if len(keyval) != 2 {
		return errors.New(`expect format key:value`)
	}
	key, val := keyval[0], keyval[1]
	(*v.header)[key] = append((*v.header)[key], val)
	return nil
}
