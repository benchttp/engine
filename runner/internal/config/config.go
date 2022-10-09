package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
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

type set map[string]struct{}

func (set set) add(values ...string) {
	for _, v := range values {
		set[v] = struct{}{}
	}
}

// Global represents the global configuration of the runner.
// It must be validated using Global.Validate before usage.
type Global struct {
	Request Request
	Runner  Runner

	Tests []tests.Case

	assignedFields set
}

// WithField returns a new Global with the input fields marked as set.
// Accepted options are limited to existing Fields, other values are
// silently ignored.
func (cfg Global) WithFields(fields ...string) Global {
	if cfg.assignedFields == nil {
		cfg.assignedFields = set{}
	}
	cfg.assignedFields.add(fields...)
	return cfg
}

// String implements fmt.Stringer. It returns an indented JSON representation
// of Config for debugging purposes.
func (cfg Global) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// Equal returns true if cfg and c are equal configurations.
func (cfg Global) Equal(c Global) bool {
	cfg.assignedFields = nil
	c.assignedFields = nil
	return reflect.DeepEqual(cfg, c)
}

// Override returns a new Config by overriding the values of base
// with the values from the Config receiver.
// Only fields previously specified by the receiver via Config.WithFields
// are replaced.
// All other values from base are preserved.
//
// The following example is equivalent to defaultConfig with the concurrency
// value from myConfig:
//
//	myConfig.
//		WithFields(FieldConcurrency).
//		Override(defaultConfig)
//
// The following example is equivalent to defaultConfig, as no field as been
// tagged via WithFields by the receiver:
//
//	myConfig.Override(defaultConfig)
func (cfg Global) Override(base Global) Global {
	for field := range cfg.assignedFields {
		switch field {
		case FieldMethod:
			base.Request.Method = cfg.Request.Method
		case FieldURL:
			base.Request.URL = cfg.Request.URL
		case FieldHeader:
			base.overrideHeader(cfg.Request.Header)
		case FieldBody:
			base.Request.Body = cfg.Request.Body
		case FieldRequests:
			base.Runner.Requests = cfg.Runner.Requests
		case FieldConcurrency:
			base.Runner.Concurrency = cfg.Runner.Concurrency
		case FieldInterval:
			base.Runner.Interval = cfg.Runner.Interval
		case FieldRequestTimeout:
			base.Runner.RequestTimeout = cfg.Runner.RequestTimeout
		case FieldGlobalTimeout:
			base.Runner.GlobalTimeout = cfg.Runner.GlobalTimeout
		case FieldTests:
			base.Tests = cfg.Tests
		}
	}
	return base
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

	if len(errs) > 0 {
		return &InvalidConfigError{errs}
	}

	return nil
}
