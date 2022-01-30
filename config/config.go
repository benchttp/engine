package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"time"
)

type Request struct {
	Method  string
	URL     *url.URL
	Timeout time.Duration
}

type RunnerOptions struct {
	Requests      int
	Concurrency   int
	GlobalTimeout time.Duration
}

// Config represents the configuration of the runner.
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

// New returns a default Config overridden with given parameters.
func New(uri string, requests, concurrency int, requestTimeout, globalTimeout time.Duration) Config {
	cfg := Config{
		Request: Request{
			Timeout: requestTimeout,
		},
		RunnerOptions: RunnerOptions{
			Requests:      requests,
			Concurrency:   concurrency,
			GlobalTimeout: globalTimeout,
		},
	}
	cfg.Request.URL, _ = url.Parse(uri) // TODO: error handling

	return MergeDefault(cfg)
}

// Default returns a default config that is safe to use.
func Default() Config {
	return defaultConfig
}

// Merge returns a Config after a base Config overridden by all non-zero values
// of override.
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
