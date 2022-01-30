package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/benchttp/runner/config"
	configfile "github.com/benchttp/runner/config/file"
	"github.com/benchttp/runner/request"
)

var (
	configFile    string
	url           string
	concurrency   int           // Number of connections to run concurrently
	requests      int           // Number of requests to run, use duration as exit condition if omitted.
	timeout       time.Duration // Timeout for each http request
	globalTimeout time.Duration // Duration of test
)

var defaultConfigFiles = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

func parseArgs() {
	flag.StringVar(&configFile, "configFile", configfile.Find(defaultConfigFiles), "Config file path")
	flag.StringVar(&url, "url", "", "Target URL to request")
	flag.IntVar(&concurrency, "concurrency", 0, "Number of connections to run concurrently")
	flag.IntVar(&requests, "requests", 0, "Number of requests to run, use duration as exit condition if omitted")
	flag.DurationVar(&timeout, "timeout", 0, "Timeout for each http request")
	flag.DurationVar(&globalTimeout, "globalTimeout", 0, "Duration of test")
	flag.Parse()
}

func main() {
	parseArgs()

	cfg := makeRunnerConfig()
	fmt.Println(cfg)

	rec := request.Do(cfg)

	fmt.Println("total:", len(rec))
}

// makeRunnerConfig retrieves a config from
func makeRunnerConfig() config.Config {
	fileConfig, err := configfile.Parse(configFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// config file is not mandatory, other errors are critical
		log.Fatal(err)
	}

	cliConfig := config.New(url, requests, concurrency, timeout, globalTimeout)

	return config.Merge(fileConfig, cliConfig)
}
