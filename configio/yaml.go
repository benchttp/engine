package configio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/benchttp/engine/benchttp"
)

// YAMLDecoder implements Decoder
type YAMLDecoder struct{ r io.Reader }

var _ decoder = (*YAMLDecoder)(nil)

// UnmarshalYAML parses the YAML-encoded data and stores the result
// in the benchttp.Runner pointed to by dst.
func UnmarshalYAML(in []byte, dst *benchttp.Runner) error {
	dec := NewYAMLDecoder(bytes.NewReader(in))
	return dec.Decode(dst)
}

func NewYAMLDecoder(r io.Reader) YAMLDecoder {
	return YAMLDecoder{r: r}
}

// Decode reads the next YAML-encoded value from its input
// and stores it in the benchttp.Runner pointed to by dst.
func (d YAMLDecoder) Decode(dst *benchttp.Runner) error {
	repr := representation{}
	if err := d.decodeRepr(&repr); err != nil {
		return err
	}
	return repr.parseAndMutate(dst)
}

// decodeRepr reads the next YAML-encoded value from its input
// and stores it in the Representation pointed to by dst.
func (d YAMLDecoder) decodeRepr(dst *representation) error {
	decoder := yaml.NewDecoder(d.r)
	decoder.KnownFields(true)
	return d.handleError(decoder.Decode(dst))
}

// handleError handles a raw yaml decoder.Decode error, filters it,
// and return the resulting error.
func (d YAMLDecoder) handleError(err error) error {
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
		if d.isCustomFieldError(msg) {
			continue
		}
		filtered.Errors = append(filtered.Errors, d.prettyErrorMessage(msg))
	}

	if len(filtered.Errors) != 0 {
		return filtered
	}

	return nil
}

// isCustomFieldError returns true if the raw error message is due
// to an allowed custom field.
func (d YAMLDecoder) isCustomFieldError(raw string) bool {
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
func (d YAMLDecoder) prettyErrorMessage(raw string) string {
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
