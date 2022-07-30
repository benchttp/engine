package configparse

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/benchttp/engine/runner"
)

// DTO is a raw data model for runner config IO.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type DTO struct {
	Extends *string `yaml:"extends" json:"extends"`

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
		Silent   *bool   `yaml:"silent" json:"silent"`
		Template *string `yaml:"template" json:"template"`
	} `yaml:"output" json:"output"`
}

// Parse parses a benchttp runner config file into a runner.ConfigGlobal
// and returns it or the first non-nil error occurring in the process,
// which can be any of the values declared in the package.
func Parse(filename string) (cfg runner.Config, err error) {
	rawConfigs, err := parseFileRecursive(filename, []DTO{}, set{})
	if err != nil {
		return
	}
	return parseAndMergeConfigs(rawConfigs)
}

// set is a collection of unique string values.
type set map[string]bool

// add adds v to the receiver. If v is already set, it returns a non-nil
// error instead.
func (s set) add(v string) error {
	if _, exists := s[v]; exists {
		return errors.New("value already set")
	}
	s[v] = true
	return nil
}

// parseFileRecursive parses a config file and its parent found from key
// "extends" recursively until the root config file is reached.
// It returns the list of all parsed configs or the first non-nil error
// occurring in the process.
func parseFileRecursive(
	filename string,
	rawConfigs []DTO,
	seen set,
) ([]DTO, error) {
	// avoid infinite recursion caused by circular reference
	if err := seen.add(filename); err != nil {
		return rawConfigs, ErrCircularExtends
	}

	// parse current file, append parsed config
	raw, err := parseFile(filename)
	if err != nil {
		return rawConfigs, err
	}
	rawConfigs = append(rawConfigs, raw)

	// root config reached: stop now and return the parsed configs
	if raw.Extends == nil {
		return rawConfigs, nil
	}

	// config has parent: resolve its path and parse it recursively
	parentPath := filepath.Join(filepath.Dir(filename), *raw.Extends)
	return parseFileRecursive(parentPath, rawConfigs, seen)
}

// parseFile parses a single config file and returns the result as an
// unmarshaledConfig and an appropriate error predeclared in the package.
func parseFile(filename string) (raw DTO, err error) {
	b, err := os.ReadFile(filename)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return raw, errWithDetails(ErrFileNotFound, filename)
	default:
		return raw, errWithDetails(ErrFileRead, filename, err)
	}

	ext := extension(filepath.Ext(filename))
	parser, err := newParser(ext)
	if err != nil {
		return raw, errWithDetails(ErrFileExt, ext, err)
	}

	if err = parser.parse(b, &raw); err != nil {
		return raw, errWithDetails(ErrParse, filename, err)
	}

	return raw, nil
}

// parseAndMergeConfigs iterates backwards over rawConfigs, parsing them
// as runner.Config and merging them into a single one.
// It returns the merged result or the first non-nil error occurring in the
// process.
func parseAndMergeConfigs(rawConfigs []DTO) (runner.Config, error) {
	globalConfig := runner.Config{}

	if len(rawConfigs) == 0 { // supposedly catched upstream, should not occur
		return globalConfig, errors.New(
			"an unacceptable error occurred parsing the config file, " +
				"please visit https://github.com/benchttp/runner/issues/new " +
				"and insult us properly",
		)
	}

	for i := len(rawConfigs) - 1; i >= 0; i-- {
		raw := rawConfigs[i]
		currentConfig, err := newParsedConfig(raw)
		if err != nil {
			return globalConfig, errWithDetails(ErrParse, err)
		}
		globalConfig = globalConfig.Override(currentConfig)
	}

	return globalConfig, nil
}

// newParsedConfig parses an input raw config as a runner.ConfigGlobal and returns
// a parsedConfig or the first non-nil error occurring in the process.
func newParsedConfig(raw DTO) (runner.Config, error) { //nolint:gocognit // acceptable complexity for a parsing func
	empty, cfg := runner.Config{}, runner.Config{}
	fieldsSet := []string{}

	markField := func(field string) {
		fieldsSet = append(fieldsSet, field)
	}

	if method := raw.Request.Method; method != nil {
		cfg.Request.Method = *method
		markField(runner.ConfigFieldMethod)
	}

	if rawURL := raw.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*raw.Request.URL, raw.Request.QueryParams)
		if err != nil {
			return empty, err
		}
		cfg.Request.URL = parsedURL
		markField(runner.ConfigFieldURL)
	}

	if header := raw.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		cfg.Request.Header = httpHeader
		markField(runner.ConfigFieldHeader)
	}

	if body := raw.Request.Body; body != nil {
		cfg.Request.Body = runner.RequestBody{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		markField(runner.ConfigFieldBody)
	}

	if requests := raw.Runner.Requests; requests != nil {
		cfg.Runner.Requests = *requests
		markField(runner.ConfigFieldRequests)
	}

	if concurrency := raw.Runner.Concurrency; concurrency != nil {
		cfg.Runner.Concurrency = *concurrency
		markField(runner.ConfigFieldConcurrency)
	}

	if interval := raw.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return empty, err
		}
		cfg.Runner.Interval = parsedInterval
		markField(runner.ConfigFieldInterval)
	}

	if requestTimeout := raw.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return empty, err
		}
		cfg.Runner.RequestTimeout = parsedTimeout
		markField(runner.ConfigFieldRequestTimeout)
	}

	if globalTimeout := raw.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return empty, err
		}
		cfg.Runner.GlobalTimeout = parsedGlobalTimeout
		markField(runner.ConfigFieldGlobalTimeout)
	}

	if silent := raw.Output.Silent; silent != nil {
		cfg.Output.Silent = *silent
		markField(runner.ConfigFieldSilent)
	}

	if template := raw.Output.Template; template != nil {
		cfg.Output.Template = *template
		markField(runner.ConfigFieldTemplate)
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
