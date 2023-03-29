package configio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/benchttp/engine/benchttp"
)

// JSONDecoder implements Decoder
type JSONDecoder struct{ r io.Reader }

var _ decoder = (*JSONDecoder)(nil)

// UnmarshalJSON parses the JSON-encoded data and stores the result
// in the benchttp.Runner pointed to by dst.
func UnmarshalJSON(in []byte, dst *benchttp.Runner) error {
	dec := NewJSONDecoder(bytes.NewReader(in))
	return dec.Decode(dst)
}

func NewJSONDecoder(r io.Reader) JSONDecoder {
	return JSONDecoder{r: r}
}

// Decode reads the next JSON-encoded value from its input
// and stores it in the benchttp.Runner pointed to by dst.
func (d JSONDecoder) Decode(dst *benchttp.Runner) error {
	repr := representation{}
	if err := d.decodeRepr(&repr); err != nil {
		return err
	}
	return repr.parseAndMutate(dst)
}

// decodeRepr reads the next JSON-encoded value from its input
// and stores it in the Representation pointed to by dst.
func (d JSONDecoder) decodeRepr(dst *representation) error {
	decoder := json.NewDecoder(d.r)
	decoder.DisallowUnknownFields()
	return d.handleError(decoder.Decode(dst))
}

// handleError handles an error from package json,
// transforms it into a user-friendly standardized format
// and returns the resulting error.
func (d JSONDecoder) handleError(err error) error {
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
	if field := d.parseUnknownFieldError(err.Error()); field != "" {
		return fmt.Errorf(`invalid field ("%s"): does not exist`, field)
	}

	return err
}

// parseJSONUnknownFieldError parses the raw string as a json error
// from an unknown field and returns the field name.
// If the raw string is not an unknown field error, it returns "".
func (d JSONDecoder) parseUnknownFieldError(raw string) (field string) {
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
