package configparse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// JSONParser implements configParser.
type JSONParser struct{}

// Parse decodes a raw JSON input in strict mode (unknown fields disallowed)
// and stores the resulting value into dst.
func (p JSONParser) Parse(in []byte, dst *Representation) error {
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.DisallowUnknownFields()
	return p.handleError(decoder.Decode(dst))
}

// handleError handle a json raw error, transforms it into a user-friendly
// standardized format and returns the resulting error.
func (p JSONParser) handleError(err error) error {
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
func (p JSONParser) parseUnknownFieldError(raw string) (field string) {
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
