package configio

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/engine/runner"
)

var (
	ErrNilConfigPtr       = errors.New("nil runner.Config pointer")
	ErrMissingUnmarshaler = errors.New("no Unmarshaler provided")
)

// Unmarshaler is the interface implemented by types that can unmarshal
// a runner.Config description of themselves.
type Unmarshaler interface {
	UnmarshalConfig(dst *runner.Config) error
}

// UnmarshalMany overrides dst iteratively with the resulting config of each
// unmarshaler, starting from the last one.
func UnmarshalMany(unmarshalers []Unmarshaler, dst *runner.Config) error {
	n := len(unmarshalers)
	if n == 0 {
		return ErrMissingUnmarshaler
	}
	if dst == nil {
		return ErrNilConfigPtr
	}

	for i := n - 1; i >= 0; i-- {
		raw := unmarshalers[i]
		cfg := runner.Config{}
		if err := raw.UnmarshalConfig(&cfg); err != nil {
			return err
		}
		*dst = cfg.Override(*dst)
	}

	return nil
}

func Unmarshal(in DTO, dst *runner.Config) error { //nolint:gocognit // acceptable complexity for a parsing func
	if dst == nil {
		return ErrNilConfigPtr
	}

	fieldsSet := make([]string, 0, len(runner.ConfigFieldsUsage))
	setField := func(field string) {
		fieldsSet = append(fieldsSet, field)
	}

	if method := in.Request.Method; method != nil {
		dst.Request.Method = *method
		setField(runner.ConfigFieldMethod)
	}

	if rawURL := in.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*in.Request.URL, in.Request.QueryParams)
		if err != nil {
			return err
		}
		dst.Request.URL = parsedURL
		setField(runner.ConfigFieldURL)
	}

	if header := in.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		dst.Request.Header = httpHeader
		setField(runner.ConfigFieldHeader)
	}

	if body := in.Request.Body; body != nil {
		dst.Request.Body = runner.RequestBody{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		setField(runner.ConfigFieldBody)
	}

	if requests := in.Runner.Requests; requests != nil {
		dst.Runner.Requests = *requests
		setField(runner.ConfigFieldRequests)
	}

	if concurrency := in.Runner.Concurrency; concurrency != nil {
		dst.Runner.Concurrency = *concurrency
		setField(runner.ConfigFieldConcurrency)
	}

	if interval := in.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return err
		}
		dst.Runner.Interval = parsedInterval
		setField(runner.ConfigFieldInterval)
	}

	if requestTimeout := in.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return err
		}
		dst.Runner.RequestTimeout = parsedTimeout
		setField(runner.ConfigFieldRequestTimeout)
	}

	if globalTimeout := in.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return err
		}
		dst.Runner.GlobalTimeout = parsedGlobalTimeout
		setField(runner.ConfigFieldGlobalTimeout)
	}

	if silent := in.Output.Silent; silent != nil {
		dst.Output.Silent = *silent
		setField(runner.ConfigFieldSilent)
	}

	if template := in.Output.Template; template != nil {
		dst.Output.Template = *template
		setField(runner.ConfigFieldTemplate)
	}

	*dst = dst.WithFields(fieldsSet...)

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
