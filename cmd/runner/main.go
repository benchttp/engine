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
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

var (
	configFile string

	uri    string
	method string // HTTP request method
	header = http.Header{}
	// HTTP body in format "type:content", type may be "raw" or "file".
	// If type is "raw", content is the data as a string. If type is "file",
	// content is the path to the file holding the data. Note: only "raw"
	// is supported at the moment.
	body config.Body

	requests       int           // Number of requests to run, use duration as exit condition if omitted.
	concurrency    int           // Number of connections to run concurrently
	interval       time.Duration // Minimum duration between two groups of requests
	requestTimeout time.Duration // Timeout for each HTTP request
	globalTimeout  time.Duration // Duration of test

	out []config.OutputStrategy // Output destinations (benchttp/json/stdin)
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
	// request method
	flag.StringVar(&method, config.FieldMethod, "", "HTTP request method")
	// request header
	flag.Var(headerValue{header: &header}, config.FieldHeader, "HTTP request header")
	// request body
	flag.Var(bodyValue{body: &body}, config.FieldBody, "HTTP request body")

	// requests number
	flag.IntVar(&requests, config.FieldRequests, 0, "Number of requests to run, use duration as exit condition if omitted")
	// concurrency
	flag.IntVar(&concurrency, config.FieldConcurrency, 0, "Number of connections to run concurrently")
	// non-conurrent requests interval
	flag.DurationVar(&interval, "interval", 0, "Minimum duration between two non concurrent requests")
	// request timeout
	flag.DurationVar(&requestTimeout, config.FieldRequestTimeout, 0, "Timeout for each HTTP request")
	// global timeout
	flag.DurationVar(&globalTimeout, config.FieldGlobalTimeout, 0, "Max duration of test")

	// output strategies
	flag.Var(outValue{out: &out}, config.FieldOut, "Output destination (benchttp/json/stdin)")

	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	parseArgs()
	fmt.Println()

	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	req, err := cfg.Request.Value()
	if err != nil {
		return err
	}

	rep, err := requester.New(requester.Config(cfg.Runner)).Run(req)
	if err != nil {
		return err
	}

	return output.New(rep, cfg).Export()
}

// parseConfig returns a config.Config initialized with config file
// options if found, overridden with CLI options.
func parseConfig() (cfg config.Global, err error) {
	fileCfg, err := configfile.Parse(configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory, other errors are critical
		return
	}

	cliCfg := config.Global{
		Request: config.Request{
			Header: header,
			Body:   body,
		}.WithURL(uri),
		Runner: config.Runner{
			Requests:       requests,
			Concurrency:    concurrency,
			Interval:       interval,
			RequestTimeout: requestTimeout,
			GlobalTimeout:  globalTimeout,
		},
		Output: config.Output{
			Out: out,
		},
	}

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
func (v headerValue) Set(raw string) error {
	keyval := strings.SplitN(raw, ":", 2)
	if len(keyval) != 2 {
		return errors.New(`expect format "<key>:<value>"`)
	}
	key, val := keyval[0], keyval[1]
	(*v.header)[key] = append((*v.header)[key], val)
	return nil
}

// bodyValue implements flag.Value
type bodyValue struct {
	body *config.Body
}

// String returns a string representation of the referenced body.
func (v bodyValue) String() string {
	return fmt.Sprint(v.body)
}

// Set reads input string in format "type:content" and sets
// the referenced body accordingly.
//
// If type is "raw", content is the data as a string.
//	"raw:{\"key\":\"value\"}" // escaped JSON
//	"raw:text" // plain text
// If type is "file", content is the path to the file holding the data.
//	"file:./path/to/file.txt"
//
// Note: only type "raw" is supported at the moment.
func (v bodyValue) Set(raw string) error {
	errFormat := fmt.Errorf(`expect format "<type>:<content>", got "%s"`, raw)

	if raw == "" {
		return errFormat
	}

	split := strings.SplitN(raw, ":", 2)
	if len(split) != 2 {
		return errFormat
	}

	btype, bcontent := split[0], split[1]
	if bcontent == "" {
		return errFormat
	}

	switch btype {
	case "raw":
		*v.body = config.NewBody(btype, bcontent)
	// case "file":
	// 	// TODO
	default:
		return fmt.Errorf(`unsupported type: %s (only "raw" accepted`, btype)
	}
	return nil
}

// outValue implements flag.Value
type outValue struct {
	out *[]config.OutputStrategy
}

// String returns a string representation of outValue.out.
func (v outValue) String() string {
	return fmt.Sprint(v.out)
}

// Set reads input string as comma-separated values and appends the values
// to the key's values of the referenced header.
func (v outValue) Set(in string) error {
	values := strings.Split(in, ",")
	if len(values) < 1 {
		return errors.New(`expect comma-separated values`)
	}
	for _, value := range values {
		*v.out = append(*v.out, config.OutputStrategy(value))
	}
	return nil
}
