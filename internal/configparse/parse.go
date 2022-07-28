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

// UnmarshaledConfig is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type UnmarshaledConfig struct {
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
	uconfs, err := parseFileRecursive(filename, []UnmarshaledConfig{}, set{})
	if err != nil {
		return
	}
	return parseAndMergeConfigs(uconfs)
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
	uconfs []UnmarshaledConfig,
	seen set,
) ([]UnmarshaledConfig, error) {
	// avoid infinite recursion caused by circular reference
	if err := seen.add(filename); err != nil {
		return uconfs, ErrCircularExtends
	}

	// parse current file, append parsed config
	uconf, err := parseFile(filename)
	if err != nil {
		return uconfs, err
	}
	uconfs = append(uconfs, uconf)

	// root config reached: stop now and return the parsed configs
	if uconf.Extends == nil {
		return uconfs, nil
	}

	// config has parent: resolve its path and parse it recursively
	parentPath := filepath.Join(filepath.Dir(filename), *uconf.Extends)
	return parseFileRecursive(parentPath, uconfs, seen)
}

// parseFile parses a single config file and returns the result as an
// unmarshaledConfig and an appropriate error predeclared in the package.
func parseFile(filename string) (uconf UnmarshaledConfig, err error) {
	b, err := os.ReadFile(filename)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return uconf, errWithDetails(ErrFileNotFound, filename)
	default:
		return uconf, errWithDetails(ErrFileRead, filename, err)
	}

	ext := extension(filepath.Ext(filename))
	parser, err := newParser(ext)
	if err != nil {
		return uconf, errWithDetails(ErrFileExt, ext, err)
	}

	if err = parser.parse(b, &uconf); err != nil {
		return uconf, errWithDetails(ErrParse, filename, err)
	}

	return uconf, nil
}

// parseAndMergeConfigs iterates backwards over uconfs, parsing them
// as runner.ConfigGlobal and merging them into a single one.
// It returns the merged result or the first non-nil error occurring in the
// process.
func parseAndMergeConfigs(uconfs []UnmarshaledConfig) (cfg runner.Config, err error) {
	if len(uconfs) == 0 { // supposedly catched upstream, should not occur
		return cfg, errors.New(
			"an unacceptable error occurred parsing the config file, " +
				"please visit https://github.com/benchttp/runner/issues/new " +
				"and insult us properly",
		)
	}

	cfg = runner.DefaultConfig()

	for i := len(uconfs) - 1; i >= 0; i-- {
		uconf := uconfs[i]
		pconf, err := newParsedConfig(uconf)
		if err != nil {
			return cfg, errWithDetails(ErrParse, err)
		}
		cfg = cfg.Override(pconf.config, pconf.fields...)
	}

	return cfg, nil
}

// parsedConfig embeds a parsed runner.ConfigGlobal and the list of its set fields.
type parsedConfig struct {
	// TODO: do not embed, use field config
	config runner.Config
	fields []string
}

// addField adds a field to the list of set fields.
func (pconf *parsedConfig) add(field string) {
	pconf.fields = append(pconf.fields, field)
}

// newParsedConfig parses an input raw config as a runner.ConfigGlobal and returns
// a parsedConfig or the first non-nil error occurring in the process.
func newParsedConfig(uconf UnmarshaledConfig) (parsedConfig, error) { //nolint:gocognit // acceptable complexity for a parsing func
	const numField = 12 // should match the number of config Fields (not critical)

	pconf := parsedConfig{
		fields: make([]string, 0, numField),
	}
	cfg := &pconf.config

	if method := uconf.Request.Method; method != nil {
		cfg.Request.Method = *method
		pconf.add(runner.ConfigFieldMethod)
	}

	if rawURL := uconf.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*uconf.Request.URL, uconf.Request.QueryParams)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Request.URL = parsedURL
		pconf.add(runner.ConfigFieldURL)
	}

	if header := uconf.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		cfg.Request.Header = httpHeader
		pconf.add(runner.ConfigFieldHeader)
	}

	if body := uconf.Request.Body; body != nil {
		cfg.Request.Body = runner.RequestBody{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		pconf.add(runner.ConfigFieldBody)
	}

	if requests := uconf.Runner.Requests; requests != nil {
		cfg.Runner.Requests = *requests
		pconf.add(runner.ConfigFieldRequests)
	}

	if concurrency := uconf.Runner.Concurrency; concurrency != nil {
		cfg.Runner.Concurrency = *concurrency
		pconf.add(runner.ConfigFieldConcurrency)
	}

	if interval := uconf.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Runner.Interval = parsedInterval
		pconf.add(runner.ConfigFieldInterval)
	}

	if requestTimeout := uconf.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Runner.RequestTimeout = parsedTimeout
		pconf.add(runner.ConfigFieldRequestTimeout)
	}

	if globalTimeout := uconf.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Runner.GlobalTimeout = parsedGlobalTimeout
		pconf.add(runner.ConfigFieldGlobalTimeout)
	}

	if silent := uconf.Output.Silent; silent != nil {
		cfg.Output.Silent = *silent
		pconf.add(runner.ConfigFieldSilent)
	}

	if template := uconf.Output.Template; template != nil {
		cfg.Output.Template = *template
		pconf.add(runner.ConfigFieldTemplate)
	}

	return pconf, nil
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
