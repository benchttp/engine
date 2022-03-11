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

// Body represents a request body associated with a type.
// The type affects the way the content is processed.
// If Type == "file", Content is read as a filepath to be resolved.
// If Type == "raw", Content is attached as-is.
//
// Note: only "raw" is supported at the moment.
type Body struct {
	Type    string
	Content []byte
}

// NewBody returns a Body initialized with the given type and content.
// For now, the only valid value for type is "raw".
func NewBody(typ, content string) Body {
	return Body{Type: typ, Content: []byte(content)}
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
	Out      []OutputStrategy
	Silent   bool
	Template string
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
		case FieldTemplate:
			cfg.Output.Template = c.Output.Template
		}
	}
	return cfg
}

// overrideHeader overrides cfg's Request.Header with the values from newHeader.
// For every key in newHeader:
//
// - If it's not present in cfg.Request.Header, it is added.
//
// - If it's already present in cfg.Request.Header, the value is replaced.
//
// - All other keys in cfg.Request.Header are left untouched.
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

	if out := cfg.Output.Out; len(out) == 0 {
		appendError(errors.New(`out: missing (want one or many of "benchttp", "json", "stdout")`))
	} else {
		for _, o := range out {
			if !IsOutput(string(o)) {
				appendError(fmt.Errorf(
					`out (%q): want one or many of "benchttp", "json", "stdout"`, o),
				)
			}
		}
	}

	if len(errs) > 0 {
		return &InvalidConfigError{errs}
	}

	return nil
}
