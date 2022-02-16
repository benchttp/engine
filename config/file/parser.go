package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

type extension string

const (
	extYML  extension = ".yml"
	extYAML extension = ".yaml"
	extJSON extension = ".json"
)

// configParser exposes a method parse to read bytes as a raw config.
type configParser interface {
	// parse parses a raw bytes input as a raw config and stores
	// the resulting value into dst.
	parse(in []byte, dst *unmarshaledConfig) error
}

// newParser returns an appropriate parser according to ext, or a non-nil
// error if ext is not an expected extension.
func newParser(ext extension) (configParser, error) {
	switch ext {
	case extYML, extYAML:
		return yamlParser{}, nil
	case extJSON:
		return jsonParser{}, nil
	default:
		return nil, errors.New("unsupported config format")
	}
}

// yamlParser implements configParser.
type yamlParser struct{}

// parse decodes a raw yaml input in strict mode (unknown fields disallowed)
// and stores the resulting value into dst.
func (p yamlParser) parse(in []byte, dst *unmarshaledConfig) error {
	decoder := yaml.NewDecoder(bytes.NewReader(in))
	decoder.KnownFields(true)
	return p.handleError(decoder.Decode(dst))
}

// handleError handles a raw yaml decoder.Decode error, filters it,
// and return the resulting error.
func (p yamlParser) handleError(err error) error {
	// yaml.TypeError errors require special handling, other errors
	// (nil included) can be returned as is.
	var typeError *yaml.TypeError
	if !errors.As(err, &typeError) {
		return err
	}

	// filter out unwanted errors
	filtered := &yaml.TypeError{}
	for _, msg := range typeError.Errors {
		// With decoder.KnownFields set to true, Decode reports any field
		// that do not match the destination structure as a non-nil error.
		// It is a wanted behavior but prevents the usage of custom aliases.
		// To work around this we allow an exception for that rule with fields
		// starting with x- (inspired by docker compose api).
		if p.isCustomFieldError(msg) {
			continue
		}
		filtered.Errors = append(filtered.Errors, p.prettyErrorMessage(msg))
	}

	if len(filtered.Errors) != 0 {
		return filtered
	}

	return nil
}

// isCustomFieldError returns true if the raw error message is due
// to an allowed custom field.
func (p yamlParser) isCustomFieldError(raw string) bool {
	customFieldRgx := regexp.MustCompile(
		// raw output example:
		// 	line 9: field x-my-alias not found in type struct { ... }
		`^line \d+: field (x-\S+) not found in type`,
	)
	return customFieldRgx.MatchString(raw)
}

// prettyErrorMessage transforms a raw Decode error message into a more
// user-friendly one by removing noisy information and returns the resulting
// value.
func (p yamlParser) prettyErrorMessage(raw string) string {
	// field not found error
	fieldNotFoundRgx := regexp.MustCompile(
		// raw output example (type unmarshaledConfig is entirely exposed):
		// 	line 11: field interval not found in type struct { ... }
		`^line (\d+): field (\S+) not found in type`,
	)
	if matches := fieldNotFoundRgx.FindStringSubmatch(raw); len(matches) >= 3 {
		line, field := matches[1], matches[2]
		return fmt.Sprintf(`line %s: invalid field ("%s"): does not exist`, line, field)
	}

	// wrong field type error
	fieldBadValueRgx := regexp.MustCompile(
		// raw output examples:
		// 	line 9: cannot unmarshal !!seq into int // unknown input value
		// 	line 10: cannot unmarshal !!str `hello` into int // known input value
		`^line (\d+): cannot unmarshal !!\w+(?: ` + "`" + `(\S+)` + "`" + `)? into (\S+)$`,
	)
	if matches := fieldBadValueRgx.FindStringSubmatch(raw); len(matches) >= 3 {
		line, value, exptype := matches[1], matches[2], matches[3]
		if value == "" {
			return fmt.Sprintf("line %s: wrong type: want %s", line, exptype)
		}
		return fmt.Sprintf(`line %s: wrong type ("%s"): want %s`, line, value, exptype)
	}

	// we may not have covered all cases, return raw output in this case
	return raw
}

// jsonParser implements configParser.
type jsonParser struct{}

// parse decodes a raw json input in strict mode (unknown fields disallowed)
// and stores the resulting value into dst.
func (p jsonParser) parse(in []byte, dst *unmarshaledConfig) error {
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.DisallowUnknownFields()
	return p.handleError(decoder.Decode(dst))
}

// handleError handle a json raw error, transforms it into a user-friendly
// standardized format and returns the resulting error.
func (p jsonParser) handleError(err error) error {
	if err == nil {
		return nil
	}

	// handle syntax error
	var errSyntax *json.SyntaxError
	if errors.As(err, &errSyntax) {
		return fmt.Errorf("syntax error near %d: %w", errSyntax.Offset, err)
	}

	// handle type error
	var errType *json.UnmarshalTypeError
	if errors.As(err, &errType) {
		return fmt.Errorf(
			"wrong type for field %s: want %s, got %s",
			errType.Field, errType.Type, errType.Value,
		)
	}

	// handle unknown field error
	if field := p.parseUnknownFieldError(err.Error()); field != "" {
		return fmt.Errorf(`invalid field ("%s"): does not exist`, field)
	}

	return err
}

// parseUnknownFieldError parses the raw string as a json error
// from an unknown field and returns the field name.
// If the raw string is not an unknown field error, it returns "".
func (p jsonParser) parseUnknownFieldError(raw string) (field string) {
	unknownFieldRgx := regexp.MustCompile(
		// raw output example:
		// 	json: unknown field "notafield"
		`json: unknown field "(\S+)"`,
	)
	if matches := unknownFieldRgx.FindStringSubmatch(raw); len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

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
}
