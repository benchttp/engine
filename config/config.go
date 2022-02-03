package config

import (
	"encoding/json"
	"fmt"
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
	// ignore err: a Config can be invalid at this point
	urlURL, _ := url.ParseRequestURI(uri)
	if urlURL == nil {
		urlURL = &url.URL{}
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

// Validate returns the config and a not nil ErrInvalid if any of the fields provided by the user is not valid
func (cfg Config) Validate() error { //nolint
	inputErrors := []error{}

	_, err := url.ParseRequestURI(cfg.Request.URL.String())
	if err != nil {
		inputErrors = append(inputErrors, fmt.Errorf("-url: %s is not a valid url", cfg.Request.URL.String()))
	}

	if cfg.RunnerOptions.Requests < 1 && cfg.RunnerOptions.Requests != -1 {
		inputErrors = append(inputErrors, fmt.Errorf("-requests: must be >= 0, we got %d", cfg.RunnerOptions.Requests))
	}

	if cfg.RunnerOptions.Concurrency < 1 && cfg.RunnerOptions.Concurrency != -1 {
		inputErrors = append(inputErrors, fmt.Errorf("-concurrency: must be > 0, we got %d", cfg.RunnerOptions.Concurrency))
	}

	if cfg.Request.Timeout < 0 {
		inputErrors = append(inputErrors, fmt.Errorf("-timeout: must be > 0, we got %d", cfg.Request.Timeout))
	}

	if cfg.RunnerOptions.GlobalTimeout < 0 {
		inputErrors = append(inputErrors, fmt.Errorf("-globalTimeout: must be > 0, we got %d", cfg.RunnerOptions.GlobalTimeout))
	}

	if len(inputErrors) > 0 {
		return &ErrInvalid{inputErrors}
	}
	return nil
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
