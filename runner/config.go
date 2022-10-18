package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/benchttp/engine/runner/internal/tests"
)

// RunnerConfig contains options relative to the runner.
type RunnerConfig struct {
	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration
}

// Config represents the global configuration of the runner.
// It must be validated using Config.Validate before usage.
type Config struct {
	Request *http.Request
	Runner  RunnerConfig
	Tests   []tests.Case
}

// String implements fmt.Stringer. It returns an indented JSON representation
// of Config for debugging purposes.
func (cfg Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// Validate returns a non-nil InvalidConfigError if any of its fields
// does not meet the requirements.
func (cfg Config) Validate() error { //nolint:gocognit
	errs := []error{}
	appendError := func(err error) {
		errs = append(errs, err)
	}

	if cfg.Request == nil {
		appendError(errors.New("unexpected nil request"))
	}

	if cfg.Runner.Requests < 1 && cfg.Runner.Requests != -1 {
		appendError(fmt.Errorf("requests (%d): want >= 0", cfg.Runner.Requests))
	}

	if cfg.Runner.Concurrency < 1 || cfg.Runner.Concurrency > cfg.Runner.Requests {
		appendError(fmt.Errorf(
			"concurrency (%d): want > 0 and <= requests (%d)",
			cfg.Runner.Concurrency, cfg.Runner.Requests,
		))
	}

	if cfg.Runner.Interval < 0 {
		appendError(fmt.Errorf("interval (%d): want >= 0", cfg.Runner.Interval))
	}

	if cfg.Runner.RequestTimeout < 1 {
		appendError(fmt.Errorf("requestTimeout (%d): want > 0", cfg.Runner.RequestTimeout))
	}

	if cfg.Runner.GlobalTimeout < 1 {
		appendError(fmt.Errorf("globalTimeout (%d): want > 0", cfg.Runner.GlobalTimeout))
	}

	if len(errs) > 0 {
		return &InvalidConfigError{errs}
	}

	return nil
}
