package conversion

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

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

func parseTests(tests []testcaseRepr) ([]benchttp.TestCase, error) {
	cases := make([]benchttp.TestCase, len(tests))
	for i, in := range tests {
		c, err := parseTestcase(in, i)
		if err != nil {
			return nil, err
		}
		cases[i] = c
	}
	return cases, nil
}

func parseTestcase(in testcaseRepr, idx int) (benchttp.TestCase, error) {
	for fieldName, fieldValue := range map[string]interface{}{
		"name":      in.Name,
		"field":     in.Field,
		"predicate": in.Predicate,
		"target":    in.Target,
	} {
		if fieldValue == nil {
			return benchttp.TestCase{}, errMissingTestField(fieldName, idx)
		}
	}

	field := benchttp.MetricsField(*in.Field)
	if err := field.Validate(); err != nil {
		return benchttp.TestCase{}, errInvalidTestField("field", idx, err)
	}

	predicate := benchttp.TestPredicate(*in.Predicate)
	if err := predicate.Validate(); err != nil {
		return benchttp.TestCase{}, errInvalidTestField("predicate", idx, err)
	}

	target, err := parseMetricValue(field, fmt.Sprint(in.Target))
	if err != nil {
		return benchttp.TestCase{}, errInvalidTestField("target", idx, err)
	}

	return benchttp.TestCase{
		Name:      *in.Name,
		Field:     field,
		Predicate: predicate,
		Target:    target,
	}, nil
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

func assertDefinedFields(fields map[string]interface{}) error {
	for name, value := range fields {
		if value == nil {
			return fmt.Errorf("%s: missing field", name)
		}
	}
	return nil
}

func errMissingTestField(fieldName string, idx int) error {
	return fmt.Errorf("tests[%d]: missing field %s", idx, fieldName)
}

func errInvalidTestField(fieldName string, idx int, err error) error {
	return fmt.Errorf("tests[%d].%s: invalid value: %w", idx, fieldName, err)
}
