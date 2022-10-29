package configio

import (
	"bytes"
	"errors"

	"github.com/benchttp/sdk/benchttp"
)

type Decoder interface {
	Decode(dst *Representation) error
	DecodeRunner(dst *benchttp.Runner) error
}

type Extension string

const (
	ExtYML  Extension = ".yml"
	ExtYAML Extension = ".yaml"
	ExtJSON Extension = ".json"
)

// DecoderOf returns the appropriate Decoder for the given extension,
// or a non-nil error if ext is not an expected extension.
func DecoderOf(ext Extension, in []byte) (Decoder, error) {
	r := bytes.NewReader(in)
	switch ext {
	case ExtYML, ExtYAML:
		return NewYAMLDecoder(r), nil
	case ExtJSON:
		return NewJSONDecoder(r), nil
	default:
		return nil, errors.New("unsupported config format")
	}
}
