package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/engine/runner/internal/tests"
)

// RequestBody represents a request body associated with a type.
// The type affects the way the content is processed.
// If Type == "file", Content is read as a filepath to be resolved.
// If Type == "raw", Content is attached as-is.
//
// Note: only "raw" is supported at the moment.
type RequestBody struct {
	Type    string
	Content []byte
}

// NewRequestBody returns a Body initialized with the given type and content.
// For now, the only valid value for type is "raw".
func NewRequestBody(typ, content string) RequestBody {
	return RequestBody{Type: typ, Content: []byte(content)}
}

// Request contains the confing options relative to a single request.
type Request struct {
	Method string
	URL    *url.URL
	Header http.Header
	Body   RequestBody
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

// Global represents the global configuration of the runner.
// It must be validated using Global.Validate before usage.
type Global struct {
	Request Request
	Runner  Runner
	Tests   []tests.Case
}

// String implements fmt.Stringer. It returns an indented JSON representation
// of Config for debugging purposes.
func (cfg Global) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// Validate returns a non-nil InvalidConfigError if any of its fields
// does not meet the requirements.
func (cfg Global) Validate() error { //nolint:gocognit
	errs := []error{}
	appendError := func(err error) {
		errs = append(errs, err)
	}

	if cfg.Request.URL == nil {
		appendError(errors.New("url: missing"))
	} else if _, err := url.ParseRequestURI(cfg.Request.URL.String()); err != nil {
		appendError(fmt.Errorf("url (%q): invalid", cfg.Request.URL.String()))
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
