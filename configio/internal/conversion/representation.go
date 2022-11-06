package conversion

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

// Repr is a raw data model for formatted runner config (json, yaml).
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
// It exposes a method Unmarshal to convert its values into a runner.Config.
type Repr struct {
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

	Tests []struct {
		Name      *string     `yaml:"name" json:"name"`
		Field     *string     `yaml:"field" json:"field"`
		Predicate *string     `yaml:"predicate" json:"predicate"`
		Target    interface{} `yaml:"target" json:"target"`
	} `yaml:"tests" json:"tests"`
}

func (repr Repr) Validate() error {
	return repr.ParseAndMutate(&benchttp.Runner{})
}

// ParseAndMutate parses the Representation receiver as a benchttp.Runner
// and stores any non-nil field value into the corresponding field
// of dst.
func (repr Repr) ParseAndMutate(dst *benchttp.Runner) error {
	if err := repr.parseRequestInto(dst); err != nil {
		return err
	}
	if err := repr.parseRunnerInto(dst); err != nil {
		return err
	}
	return repr.parseTestsInto(dst)
}

func (repr Repr) parseRequestInto(dst *benchttp.Runner) error {
	if dst.Request == nil {
		dst.Request = &http.Request{}
	}

	if method := repr.Request.Method; method != nil {
		dst.Request.Method = *method
	}

	if rawURL := repr.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*rawURL, repr.Request.QueryParams)
		if err != nil {
			return fmt.Errorf(`configio: invalid url: %q`, *rawURL)
		}
		dst.Request.URL = parsedURL
	}

	if header := repr.Request.Header; len(header) != 0 {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		dst.Request.Header = httpHeader
	}

	if body := repr.Request.Body; body != nil {
		switch body.Type {
		case "raw":
			dst.Request.Body = io.NopCloser(bytes.NewReader([]byte(body.Content)))
		default:
			return errors.New(`configio: request.body.type: only "raw" accepted`)
		}
	}

	return nil
}

func (repr Repr) parseRunnerInto(dst *benchttp.Runner) error {
	if requests := repr.Runner.Requests; requests != nil {
		dst.Requests = *requests
	}

	if concurrency := repr.Runner.Concurrency; concurrency != nil {
		dst.Concurrency = *concurrency
	}

	if interval := repr.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return err
		}
		dst.Interval = parsedInterval
	}

	if requestTimeout := repr.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return err
		}
		dst.RequestTimeout = parsedTimeout
	}

	if globalTimeout := repr.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return err
		}
		dst.GlobalTimeout = parsedGlobalTimeout
	}

	return nil
}

func (repr Repr) parseTestsInto(dst *benchttp.Runner) error {
	testSuite := repr.Tests
	if len(testSuite) == 0 {
		return nil
	}

	cases := make([]benchttp.TestCase, len(testSuite))
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
			return err
		}

		field := benchttp.MetricsField(*t.Field)
		if err := field.Validate(); err != nil {
			return fmt.Errorf("%s: %s", fieldPath("field"), err)
		}

		predicate := benchttp.TestPredicate(*t.Predicate)
		if err := predicate.Validate(); err != nil {
			return fmt.Errorf("%s: %s", fieldPath("predicate"), err)
		}

		target, err := parseMetricValue(field, fmt.Sprint(t.Target))
		if err != nil {
			return fmt.Errorf("%s: %s", fieldPath("target"), err)
		}

		cases[i] = benchttp.TestCase{
			Name:      *t.Name,
			Field:     field,
			Predicate: predicate,
			Target:    target,
		}
	}

	dst.Tests = cases
	return nil
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
	field benchttp.MetricsField,
	inputValue string,
) (benchttp.MetricsValue, error) {
	fieldType := field.Type()
	handleError := func(v interface{}, err error) (benchttp.MetricsValue, error) {
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

type Reprs []Repr

// MergeInto successively parses the given representations into dst.
//
// The input Representation slice must never be nil or empty, otherwise it panics.
func (reprs Reprs) MergeInto(dst *benchttp.Runner) error {
	if len(reprs) == 0 { // supposedly catched upstream, should not occur
		panicInternal("parseAndMergeReprs", "nil or empty []Representation")
	}

	for _, repr := range reprs {
		if err := repr.ParseAndMutate(dst); err != nil {
			return err
			// TODO: uncomment once wrapped from configio/file.go
			// return errorutil.WithDetails(ErrFileParse, err)
		}
	}
	return nil
}

func panicInternal(funcname, detail string) {
	const reportURL = "https://github.com/benchttp/sdk/issues/new"
	source := fmt.Sprintf("configio.%s", funcname)
	panic(fmt.Sprintf(
		"%s: unexpected internal error: %s, please file an issue at %s",
		source, detail, reportURL,
	))
}
