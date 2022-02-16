package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	Method  string
	URL     *url.URL
	Header  http.Header
	Timeout time.Duration
	Body    Body
}

// RunnerOptions contains options relative to the runner.
type RunnerOptions struct {
	Requests      int
	Concurrency   int
	Interval      time.Duration
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
	if cfg.Request.URL == nil {
		return nil, errors.New("empty url")
	}
	rawURL := cfg.Request.URL.String()
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return nil, errors.New("bad url")
	}

	req, err := http.NewRequest(cfg.Request.Method, rawURL, bytes.NewReader(cfg.Request.Body.Content))
	if err != nil {
		return nil, err
	}
	req.Header = cfg.Request.Header
	return req, nil
}

// Override returns a new Config based on cfg with overridden values from c.
// Only fields specified in options are replaced. Accepted options are limited
// to existing Fields, other values are silently ignored.
func (cfg Config) Override(c Config, fields ...string) Config {
	for _, field := range fields {
		switch field {
		case FieldMethod:
			cfg.Request.Method = c.Request.Method
		case FieldURL:
			cfg.Request.URL = c.Request.URL
		case FieldHeader:
			cfg.overrideHeader(c.Request.Header)
		case FieldTimeout:
			cfg.Request.Timeout = c.Request.Timeout
		case FieldRequests:
			cfg.RunnerOptions.Requests = c.RunnerOptions.Requests
		case FieldConcurrency:
			cfg.RunnerOptions.Concurrency = c.RunnerOptions.Concurrency
		case FieldInterval:
			cfg.RunnerOptions.Interval = c.RunnerOptions.Interval
		case FieldGlobalTimeout:
			cfg.RunnerOptions.GlobalTimeout = c.RunnerOptions.GlobalTimeout
		case FieldBody:
			cfg.Request.Body = c.Request.Body
		}
	}
	return cfg
}

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

// WithURL sets the current Config to the parsed *url.URL from rawURL
// and returns it. Any errors is discarded as a Config can be invalid
// until Config.Validate is called. The url is guaranteed not to be nil.
func (cfg Config) WithURL(rawURL string) Config {
	// ignore err: a Config can be invalid at this point
	urlURL, _ := url.ParseRequestURI(rawURL)
	if urlURL == nil {
		urlURL = &url.URL{}
	}
	cfg.Request.URL = urlURL
	return cfg
}

// Validate returns the config and a not nil ErrInvalid if any of the fields provided by the user is not valid
func (cfg Config) Validate() error { //nolint:gocognit
	inputErrors := []error{}

	if cfg.Request.URL == nil {
		inputErrors = append(inputErrors, errors.New("-url: missing url"))
	} else if _, err := url.ParseRequestURI(cfg.Request.URL.String()); err != nil {
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

	if cfg.RunnerOptions.Interval < 0 {
		inputErrors = append(inputErrors, fmt.Errorf("-interval: must be > 0, we got %d", cfg.RunnerOptions.Interval))
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

// ParseBodyContent parses raw and returns the content as a string or an error.
// raw is in format "type:content", where type may be "raw" or "file".
//
// If type is "raw", content is the data as a string.
//	"raw:{\"key\":\"value\"}" // escaped JSON
//	"raw:text" // plain text
// If type is "file", content is the path to the file holding the data.
//	"file:./path/to/file.txt"
//
// Note: only type "raw" is supported at the moment.
func ParseBody(raw string) (Body, error) {
	if raw == "" {
		// Body is nil.
		return Body{}, nil
	}

	split := strings.SplitN(raw, ":", 2)
	if len(split) != 2 {
		return Body{}, fmt.Errorf("expected format \"<type>:<content>\", got %s", raw)
	}
	if split[1] == "" {
		return Body{}, errors.New("got type but no content")
	}

	switch split[0] {
	case "raw":
		return NewBody("raw", split[1]), nil
	// case "file":
	// 	// TODO
	default:
		return Body{}, fmt.Errorf("unsupported type %s", split[0])
	}
}
