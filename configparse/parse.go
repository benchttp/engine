package configparse

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/benchttp/engine/runner"
)

// Representation is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type Representation struct {
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
		Silent *bool `yaml:"silent" json:"silent"`
	} `yaml:"output" json:"output"`

	Tests []struct {
		Name      *string     `yaml:"name" json:"name"`
		Field     *string     `yaml:"field" json:"field"`
		Predicate *string     `yaml:"predicate" json:"predicate"`
		Target    interface{} `yaml:"target" json:"target"`
	} `yaml:"tests" json:"tests"`
}

// ParseRepresentation parses an input raw config as a runner.ConfigGlobal and returns
// a parsed Config or the first non-nil error occurring in the process.
func ParseRepresentation(repr Representation) (runner.Config, error) { //nolint:gocognit // acceptable complexity for a parsing func
	cfg := runner.Config{}
	fieldsSet := []string{}

	addField := func(field string) {
		fieldsSet = append(fieldsSet, field)
	}

	abort := func(err error) (runner.Config, error) {
		return runner.Config{}, err
	}

	if method := repr.Request.Method; method != nil {
		cfg.Request.Method = *method
		addField(runner.ConfigFieldMethod)
	}

	if rawURL := repr.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*repr.Request.URL, repr.Request.QueryParams)
		if err != nil {
			return abort(err)
		}
		cfg.Request.URL = parsedURL
		addField(runner.ConfigFieldURL)
	}

	if header := repr.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		cfg.Request.Header = httpHeader
		addField(runner.ConfigFieldHeader)
	}

	if body := repr.Request.Body; body != nil {
		cfg.Request.Body = runner.RequestBody{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		addField(runner.ConfigFieldBody)
	}

	if requests := repr.Runner.Requests; requests != nil {
		cfg.Runner.Requests = *requests
		addField(runner.ConfigFieldRequests)
	}

	if concurrency := repr.Runner.Concurrency; concurrency != nil {
		cfg.Runner.Concurrency = *concurrency
		addField(runner.ConfigFieldConcurrency)
	}

	if interval := repr.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return abort(err)
		}
		cfg.Runner.Interval = parsedInterval
		addField(runner.ConfigFieldInterval)
	}

	if requestTimeout := repr.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return abort(err)
		}
		cfg.Runner.RequestTimeout = parsedTimeout
		addField(runner.ConfigFieldRequestTimeout)
	}

	if globalTimeout := repr.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return abort(err)
		}
		cfg.Runner.GlobalTimeout = parsedGlobalTimeout
		addField(runner.ConfigFieldGlobalTimeout)
	}

	if silent := repr.Output.Silent; silent != nil {
		cfg.Output.Silent = *silent
		addField(runner.ConfigFieldSilent)
	}

	testSuite := repr.Tests
	if len(testSuite) == 0 {
		return cfg.WithFields(fieldsSet...), nil
	}

	cases := make([]runner.TestCase, len(testSuite))
	for i, t := range testSuite {
		fieldPath := func(caseField string) string {
			return fmt.Sprintf("tests[%d].%s", i, caseField)
		}

		if err := requireConfigFields(map[string]interface{}{
			fieldPath("name"):      t.Name,
			fieldPath("field"):     t.Field,
			fieldPath("predicate"): t.Predicate,
			fieldPath("target"):    t.Target,
		}); err != nil {
			return abort(err)
		}

		field := runner.MetricsField(*t.Field)
		if err := field.Validate(); err != nil {
			return abort(fmt.Errorf("%s: %s", fieldPath("field"), err))
		}

		predicate := runner.TestPredicate(*t.Predicate)
		if err := predicate.Validate(); err != nil {
			return abort(fmt.Errorf("%s: %s", fieldPath("predicate"), err))
		}

		target, err := parseMetricValue(field, fmt.Sprint(t.Target))
		if err != nil {
			return abort(fmt.Errorf("%s: %s", fieldPath("target"), err))
		}

		cases[i] = runner.TestCase{
			Name:      *t.Name,
			Field:     field,
			Predicate: predicate,
			Target:    target,
		}
	}
	cfg.Tests = cases
	addField(runner.ConfigFieldTests)

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

func parseMetricValue(
	field runner.MetricsField,
	inputValue string,
) (runner.MetricsValue, error) {
	fieldType := field.Type()
	handleError := func(v interface{}, err error) (runner.MetricsValue, error) {
		if err != nil {
			return nil, fmt.Errorf(
				"value %q is incompatible with field %s (want %s)",
				inputValue, field, fieldType,
			)
		}
		return v, nil
	}
	switch fieldType {
	case "int":
		return handleError(strconv.Atoi(inputValue))
	case "time.Duration":
		return handleError(time.ParseDuration(inputValue))
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

func requireConfigFields(fields map[string]interface{}) error {
	for name, value := range fields {
		if value == nil {
			return fmt.Errorf("%s: missing field", name)
		}
	}
	return nil
}
