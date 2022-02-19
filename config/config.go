package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Body struct {
	Type    string
	Content []byte
}

// // To return a Body pbject with Body.Content as a string
// func (b Body) String() string {
// 	bodyObject := "\"Body\": "
// 	bodyType := "\"Type\" :\"" + b.Type + "\""
// 	bodyContent := "\"Content\": \"" + string(b.Content) + "\""
// 	return fmt.Sprintf("{%s\r\t%s\r\t%s\r}", bodyObject, bodyType, bodyContent)
// }

func NewBody(bodyType, bodyContent string) Body {
	var body Body
	body.Type = bodyType
	body.Content = []byte(bodyContent)
	return body
}

// Request contains the confing options relative to a single request.
type Request struct {
	Method string
	URL    *url.URL
	Header http.Header
	Body   Body
}

// Value generates a *http.Request based on Request and returns it
// or any non-nil error that occurred.
func (r Request) Value() (*http.Request, error) {
	if r.URL == nil {
		return nil, errors.New("empty url")
	}
	rawURL := r.URL.String()
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return nil, errors.New("bad url")
	}

	req, err := http.NewRequest(r.Method, rawURL, bytes.NewReader(r.Body.Content))
	if err != nil {
		return nil, err
	}
	req.Header = r.Header
	return req, nil
}

// WithURL sets the current Request with the parsed *url.URL from rawURL
// and returns it. Any errors is discarded as a Config can be invalid
// until Config.Validate is called. The url is always non-nil.
func (r Request) WithURL(rawURL string) Request {
	// ignore err: a Config can be invalid at this point
	urlURL, _ := url.ParseRequestURI(rawURL)
	if urlURL == nil {
		urlURL = &url.URL{}
	}
	r.URL = urlURL
	return r
}

// Runner contains options relative to the runner.
type Runner struct {
	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration
}

// Output contains options relative to the output.
type Output struct {
	Out    []OutputStrategy
	Silent bool
}

// Global represents the global configuration of the runner.
// It must be validated using Global.Validate before usage.
type Global struct {
	Request Request
	Runner  Runner
	Output  Output
}

// String returns an indented JSON representation of Config
// for debugging purposes.
func (cfg Global) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// Override returns a new Config based on cfg with overridden values from c.
// Only fields specified in options are replaced. Accepted options are limited
// to existing Fields, other values are silently ignored.
func (cfg Global) Override(c Global, fields ...string) Global {
	for _, field := range fields {
		switch field {
		case FieldMethod:
			cfg.Request.Method = c.Request.Method
		case FieldURL:
			cfg.Request.URL = c.Request.URL
		case FieldHeader:
			cfg.overrideHeader(c.Request.Header)
		case FieldBody:
			cfg.Request.Body = c.Request.Body
		case FieldRequests:
			cfg.Runner.Requests = c.Runner.Requests
		case FieldConcurrency:
			cfg.Runner.Concurrency = c.Runner.Concurrency
		case FieldInterval:
			cfg.Runner.Interval = c.Runner.Interval
		case FieldRequestTimeout:
			cfg.Runner.RequestTimeout = c.Runner.RequestTimeout
		case FieldGlobalTimeout:
			cfg.Runner.GlobalTimeout = c.Runner.GlobalTimeout
		case FieldOut:
			cfg.Output.Out = c.Output.Out
		case FieldSilent:
			cfg.Output.Silent = c.Output.Silent
		}
	}
	return cfg
}

func (cfg *Global) overrideHeader(newHeader http.Header) {
	if newHeader == nil {
		return
	}
	if cfg.Request.Header == nil {
		cfg.Request.Header = http.Header{}
	}
	for k, v := range newHeader {
		cfg.Request.Header[k] = v
	}
}

// Validate returns the config and a not nil ErrInvalid if any of the fields provided by the user is not valid
func (cfg Global) Validate() error { //nolint:gocognit
	inputErrors := []error{}

	appendError := func(err error) {
		inputErrors = append(inputErrors, err)
	}

	if cfg.Request.URL == nil {
		appendError(errors.New("-url: missing url"))
	} else if _, err := url.ParseRequestURI(cfg.Request.URL.String()); err != nil {
		appendError(fmt.Errorf("-url: %s is not a valid url", cfg.Request.URL.String()))
	}

	if cfg.Runner.Requests < 1 && cfg.Runner.Requests != -1 {
		appendError(fmt.Errorf("-requests: must be >= 0, we got %d", cfg.Runner.Requests))
	}

	if cfg.Runner.Concurrency < 1 && cfg.Runner.Concurrency != -1 {
		appendError(fmt.Errorf("-concurrency: must be > 0, we got %d", cfg.Runner.Concurrency))
	}

	if cfg.Runner.Interval < 0 {
		appendError(fmt.Errorf("-interval: must be > 0, we got %d", cfg.Runner.Interval))
	}

	if cfg.Runner.RequestTimeout < 0 {
		appendError(fmt.Errorf("-timeout: must be > 0, we got %d", cfg.Runner.RequestTimeout))
	}

	if cfg.Runner.GlobalTimeout < 0 {
		appendError(fmt.Errorf("-globalTimeout: must be > 0, we got %d", cfg.Runner.GlobalTimeout))
	}

	if out := cfg.Output.Out; len(out) == 0 {
		appendError(errors.New(`-out: missing (want one or many of "benchttp", "json", "stdin")`))
	} else {
		for _, o := range out {
			if !IsOutput(string(o)) {
				appendError(fmt.Errorf(
					`-out: invalid value: %s (want one or many of "benchttp", "json", "stdin")`, o),
				)
			}
		}
	}

	if len(inputErrors) > 0 {
		return &ErrInvalid{inputErrors}
	}

	return nil
}

// Default returns a default config that is safe to use.
func Default() Global {
	return defaultConfig
}
