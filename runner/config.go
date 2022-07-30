package runner

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"
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

// RequestConfig contains the confing options relative to a single request.
type RequestConfig struct {
	Method string
	URL    *url.URL
	Header http.Header
	Body   RequestBody
}

// Value generates a *http.Request based on Request and returns it
// or any non-nil error that occurred.
func (r RequestConfig) Value() (*http.Request, error) {
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
func (r RequestConfig) WithURL(rawURL string) RequestConfig {
	// ignore err: a Config can be invalid at this point
	urlURL, _ := url.ParseRequestURI(rawURL)
	if urlURL == nil {
		urlURL = &url.URL{}
	}
	r.URL = urlURL
	return r
}

// RecorderConfig contains options relative to the runner.
type RecorderConfig struct {
	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration
}

// OutputConfig contains options relative to the output.
type OutputConfig struct {
	Silent   bool
	Template string
}

type set map[string]struct{}

// Config represents the global configuration of the runner.
// It must be validated using Config.Validate before usage.
type Config struct {
	Request RequestConfig
	Runner  RecorderConfig
	Output  OutputConfig

	fieldsSet set
}

// WithField returns a new Global with the input fields marked as set.
func (cfg Config) WithFields(fields ...string) Config {
	fieldsSet := cfg.fieldsSet
	if fieldsSet == nil {
		fieldsSet = set{}
	}
	for _, field := range fields {
		fieldsSet[field] = struct{}{}
	}
	cfg.fieldsSet = fieldsSet
	return cfg
}

// Equal returns true if cfg and c are equal configurations.
func (cfg Config) Equal(c Config) bool {
	cfg.fieldsSet = nil
	c.fieldsSet = nil
	return reflect.DeepEqual(cfg, c)
}

// Override returns a new Config based on cfg with overridden values from c.
// Only fields specified in options are replaced. Accepted options are limited
// to existing Fields, other values are silently ignored.
func (cfg Config) Override(c Config) Config {
	if len(cfg.fieldsSet) == 0 {
		return c
	}
	for field := range cfg.fieldsSet {
		switch field {
		case ConfigFieldMethod:
			c.Request.Method = cfg.Request.Method
		case ConfigFieldURL:
			c.Request.URL = cfg.Request.URL
		case ConfigFieldHeader:
			c.overrideHeader(cfg.Request.Header)
		case ConfigFieldBody:
			c.Request.Body = cfg.Request.Body
		case ConfigFieldRequests:
			c.Runner.Requests = cfg.Runner.Requests
		case ConfigFieldConcurrency:
			c.Runner.Concurrency = cfg.Runner.Concurrency
		case ConfigFieldInterval:
			c.Runner.Interval = cfg.Runner.Interval
		case ConfigFieldRequestTimeout:
			c.Runner.RequestTimeout = cfg.Runner.RequestTimeout
		case ConfigFieldGlobalTimeout:
			c.Runner.GlobalTimeout = cfg.Runner.GlobalTimeout
		case ConfigFieldSilent:
			c.Output.Silent = cfg.Output.Silent
		case ConfigFieldTemplate:
			c.Output.Template = cfg.Output.Template
		}
	}
	return c
}

// overrideHeader overrides cfg's Request.Header with the values from newHeader.
// For every key in newHeader:
//
// - If it's not present in cfg.Request.Header, it is added.
//
// - If it's already present in cfg.Request.Header, the value is replaced.
//
// - All other keys in cfg.Request.Header are left untouched.
func (cfg *Config) overrideHeader(newHeader http.Header) {
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
func (cfg Config) Validate() error { //nolint:gocognit
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

var defaultConfig = Config{
	Request: RequestConfig{
		Method: "GET",
		URL:    &url.URL{},
		Header: http.Header{},
		Body:   RequestBody{},
	},
	Runner: RecorderConfig{
		Concurrency:    10,
		Requests:       100,
		Interval:       0 * time.Second,
		RequestTimeout: 5 * time.Second,
		GlobalTimeout:  30 * time.Second,
	},
	Output: OutputConfig{
		Silent:   false,
		Template: "",
	},
}

// Default returns a default config that is safe to use.
func DefaultConfig() Config {
	return defaultConfig
}
