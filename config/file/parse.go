package file

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/benchttp/runner/config"
)

// unmarshaledConfig is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type unmarshaledConfig struct {
	Request struct {
		Method      *string             `yaml:"method" json:"method"`
		URL         *string             `yaml:"url" json:"url"`
		QueryParams map[string]string   `yaml:"queryParams" json:"queryParams"`
		Header      map[string][]string `yaml:"header" json:"header"`
		Body        *struct {
			Type    string `yaml:"type" json:"type"`
			Content string `yaml:"content" json:"content"`
		} `yaml:"body" json:"body"`
	} `yaml:"request" json:"request"`

	Runner struct {
		Requests       *int    `yaml:"requests" json:"requests"`
		Concurrency    *int    `yaml:"concurrency" json:"concurrency"`
		Interval       *string `yaml:"interval" json:"interval"`
		RequestTimeout *string `yaml:"requestTimeout" json:"requestTimeout"`
		GlobalTimeout  *string `yaml:"globalTimeout" json:"globalTimeout"`
	} `yaml:"runner" json:"runner"`

	Output struct {
		Out      *[]string `yaml:"out" json:"out"`
		Silent   *bool     `yaml:"silent" json:"silent"`
		Template *string   `yaml:"template" json:"template"`
	} `yaml:"output" json:"output"`
}

// Parse parses a benchttp runner config file into a config.Config
// and returns it or the first non-nil error occurring in the process.
func Parse(cfgpath string) (cfg config.Global, err error) {
	b, err := os.ReadFile(cfgpath)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return cfg, errWithDetails(ErrFileNotFound, cfgpath)
	default:
		return cfg, errWithDetails(ErrFileRead, cfgpath, err)
	}

	ext := extension(path.Ext(cfgpath))
	parser, err := newParser(ext)
	if err != nil {
		return cfg, errWithDetails(ErrFileExt, ext, err)
	}

	var rawCfg unmarshaledConfig
	if err = parser.parse(b, &rawCfg); err != nil {
		return cfg, errWithDetails(ErrParse, cfgpath, err)
	}

	cfg, err = parseRawConfig(rawCfg)
	if err != nil {
		return cfg, errWithDetails(ErrParse, cfgpath, err)
	}

	return
}

// parseRawConfig parses an input raw config as a config.Config and returns it
// or the first non-nil error occurring in the process.
func parseRawConfig(raw unmarshaledConfig) (config.Global, error) { //nolint:gocognit // acceptable complexity for a parsing func
	const numField = 10 // should match the number of config Fields (not critical)

	cfg := config.Global{}
	fields := make([]string, 0, numField)

	appendField := func(field string) {
		fields = append(fields, field)
	}

	if method := raw.Request.Method; method != nil {
		cfg.Request.Method = *method
		appendField(config.FieldMethod)
	}

	if rawURL := raw.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*raw.Request.URL, raw.Request.QueryParams)
		if err != nil {
			return config.Global{}, err
		}
		cfg.Request.URL = parsedURL
		appendField(config.FieldURL)
	}

	if header := raw.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		cfg.Request.Header = httpHeader
		appendField(config.FieldHeader)
	}

	if body := raw.Request.Body; body != nil {
		cfg.Request.Body = config.Body{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		fields = append(fields, config.FieldBody)
	}

	if requests := raw.Runner.Requests; requests != nil {
		cfg.Runner.Requests = *requests
		appendField(config.FieldRequests)
	}

	if concurrency := raw.Runner.Concurrency; concurrency != nil {
		cfg.Runner.Concurrency = *concurrency
		appendField(config.FieldConcurrency)
	}

	if interval := raw.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return config.Global{}, err
		}
		cfg.Runner.Interval = parsedInterval
		appendField(config.FieldInterval)
	}

	if requestTimeout := raw.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return config.Global{}, err
		}
		cfg.Runner.RequestTimeout = parsedTimeout
		appendField(config.FieldRequestTimeout)
	}

	if globalTimeout := raw.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return config.Global{}, err
		}
		cfg.Runner.GlobalTimeout = parsedGlobalTimeout
		appendField(config.FieldGlobalTimeout)
	}

	if outs := raw.Output.Out; outs != nil {
		for _, out := range *outs {
			cfg.Output.Out = append(cfg.Output.Out, config.OutputStrategy(out))
		}
		appendField(config.FieldOut)
	}

	if silent := raw.Output.Silent; silent != nil {
		cfg.Output.Silent = *silent
		appendField(config.FieldSilent)
	}

	if template := raw.Output.Template; template != nil {
		cfg.Output.Template = *template
		appendField(config.FieldTemplate)
	}

	return config.Default().Override(cfg, fields...), nil
}

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
