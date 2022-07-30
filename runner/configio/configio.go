package configio

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/engine/runner"
)

// Interface exposes a method Config to retrieve a runner.Config.
// Its main purpose is to be able to use any slice of struct that
// implements this interface (notably: structs that embed DTO)
// as a valid argument for ParseMany without the need of a conversion.
type Interface interface {
	Config() (runner.Config, error)
}

// ParseMany parses and overrides raws iteratively, from right to left,
// starting with a zero-value runner.Config.
// It returns the resulting merged runner.Config or the first non-nil error
// occurring in the process.
func ParseMany(raws []Interface) (runner.Config, error) {
	return parseMany(raws, runner.Config{})
}

// ParseManyWithDefault does the same as ParseMany, but uses the default
// runner.Config as the base.
func ParseManyWithDefault(raws []Interface) (runner.Config, error) {
	return parseMany(raws, runner.DefaultConfig())
}

// Parse turns a DTO into a runner.Config and returns it or the first
// non-nil error occurring in the process.
func Parse(parser DTO) (runner.Config, error) {
	return parseSingle(parser, runner.Config{})
}

// Parse turns a DTO into a runner.Config, merges it with a the default config
// and returns it or the first non-nil error occurring in the process.
func ParseWithDefault(parser DTO) (runner.Config, error) {
	return parseSingle(parser, runner.DefaultConfig())
}

func parseMany(raws []Interface, baseConfig runner.Config) (runner.Config, error) {
	if len(raws) == 0 {
		return baseConfig, errors.New("no configs provided")
	}

	merged := baseConfig
	for i := len(raws) - 1; i >= 0; i-- {
		raw := raws[i]
		currentConfig, err := raw.Config()
		if err != nil {
			return merged, err
		}
		merged = currentConfig.Override(merged)
	}

	return merged, nil
}

func parseSingle(parser DTO, baseConfig runner.Config) (runner.Config, error) { //nolint:gocognit // acceptable complexity for a parsing func
	cfg := baseConfig
	fieldsSet := []string{}

	setField := func(field string) {
		fieldsSet = append(fieldsSet, field)
	}

	if method := parser.Request.Method; method != nil {
		cfg.Request.Method = *method
		setField(runner.ConfigFieldMethod)
	}

	if rawURL := parser.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*parser.Request.URL, parser.Request.QueryParams)
		if err != nil {
			return runner.Config{}, err
		}
		cfg.Request.URL = parsedURL
		setField(runner.ConfigFieldURL)
	}

	if header := parser.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		cfg.Request.Header = httpHeader
		setField(runner.ConfigFieldHeader)
	}

	if body := parser.Request.Body; body != nil {
		cfg.Request.Body = runner.RequestBody{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		setField(runner.ConfigFieldBody)
	}

	if requests := parser.Runner.Requests; requests != nil {
		cfg.Runner.Requests = *requests
		setField(runner.ConfigFieldRequests)
	}

	if concurrency := parser.Runner.Concurrency; concurrency != nil {
		cfg.Runner.Concurrency = *concurrency
		setField(runner.ConfigFieldConcurrency)
	}

	if interval := parser.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return runner.Config{}, err
		}
		cfg.Runner.Interval = parsedInterval
		setField(runner.ConfigFieldInterval)
	}

	if requestTimeout := parser.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return runner.Config{}, err
		}
		cfg.Runner.RequestTimeout = parsedTimeout
		setField(runner.ConfigFieldRequestTimeout)
	}

	if globalTimeout := parser.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return runner.Config{}, err
		}
		cfg.Runner.GlobalTimeout = parsedGlobalTimeout
		setField(runner.ConfigFieldGlobalTimeout)
	}

	if silent := parser.Output.Silent; silent != nil {
		cfg.Output.Silent = *silent
		setField(runner.ConfigFieldSilent)
	}

	if template := parser.Output.Template; template != nil {
		cfg.Output.Template = *template
		setField(runner.ConfigFieldTemplate)
	}

	return cfg.WithFields(fieldsSet...), nil
}

// helpers

// parseAndBuildURL parses a raw string as a *url.URL and adds any extra
// query parameters. It returns the first non-nil error occurring in the
// process.
func parseAndBuildURL(raw string, qp map[string]string) (*url.URL, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, err
	}

	// retrieve url query, add extra params, re-attach to url
	if qp != nil {
		q := u.Query()
		for k, v := range qp {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u, nil
}

// parseOptionalDuration parses the raw string as a time.Duration
// and returns the parsed value or a non-nil error.
// Contrary to time.ParseDuration, it does not return an error
// if raw == "".
func parseOptionalDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}
