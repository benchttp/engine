package config

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// Request contains the confing options relative to a single request.
type Request struct {
	Method  string
	URL     *url.URL
	Timeout time.Duration
}

// RunnerOptions contains options relative to the runner.
type RunnerOptions struct {
	Requests      int
	Concurrency   int
	GlobalTimeout time.Duration
}

// Config represents the configuration of the runner.
// It must be validated using Config.Validate before usage.
type Config struct {
	Request       Request
	RunnerOptions RunnerOptions
}

// String returns an indented JSON representation of Config
// for debugging purposes.
func (cfg Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// HTTPRequest returns a *http.Request created from Target. Returns any non-nil
// error that occurred.
func (cfg Config) HTTPRequest() (*http.Request, error) {
	return http.NewRequest(
		cfg.Request.Method,
		cfg.Request.URL.String(),
		nil, // TODO: handle body
	)
}

// New returns a Config initialized with given parameters. The returned Config
// is not guaranteed to be safe: it must be validated using Config.Validate
// before usage.
func New(uri string, requests, concurrency int, requestTimeout, globalTimeout time.Duration) Config {
	var urlURL *url.URL
	if uri != "" {
		// ignore err: a Config can be invalid at this point
		urlURL, _ = url.Parse(uri)
	}
	return Config{
		Request: Request{
			URL:     urlURL,
			Timeout: requestTimeout,
		},
		RunnerOptions: RunnerOptions{
			Requests:      requests,
			Concurrency:   concurrency,
			GlobalTimeout: globalTimeout,
		},
	}
}

// Default returns a default config that is safe to use.
func Default() Config {
	return defaultConfig
}

// Merge returns a Config after a base Config overridden by all non-zero values
// of override. The returned Config is not guaranteed to be safe: it must be
// validated using Config.Validate before usage.
func Merge(base, override Config) Config {
	if override.Request.Method != "" {
		base.Request.Method = override.Request.Method
	}
	newURL := override.Request.URL
	if newURL != nil && newURL.String() != "" {
		base.Request.URL = override.Request.URL
	}
	if override.Request.Timeout != 0 {
		base.Request.Timeout = override.Request.Timeout
	}
	if override.RunnerOptions.Requests != 0 {
		base.RunnerOptions.Requests = override.RunnerOptions.Requests
	}
	if override.RunnerOptions.Concurrency != 0 {
		base.RunnerOptions.Concurrency = override.RunnerOptions.Concurrency
	}
	if override.RunnerOptions.GlobalTimeout != 0 {
		base.RunnerOptions.GlobalTimeout = override.RunnerOptions.GlobalTimeout
	}
	return base
}

// MergeDefault merges override with the default config calling Merge.
// The returned Config is not guaranteed to be safe: it must be validated
// using Config.Validate before usage.
func MergeDefault(override Config) Config {
	return Merge(Default(), override)
}

// Validate returns an unimplemented error.
//
// Once implemented, Validate will return ErrInvalid if any of its fields
// does not meet the runner requirements.
//
// TODO: https://github.com/benchttp/runner/issues/20
func (cfg Config) Validate() error {
	return errors.New("unimplemented")
}
