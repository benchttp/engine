package configio

import (
	"bytes"
	"fmt"

	"github.com/benchttp/sdk/benchttp"
)

type Decoder interface {
	Decode(dst *Representation) error
	DecodeRunner(dst *benchttp.Runner) error
}

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// DecoderOf returns the appropriate Decoder for the given Format.
// It panics if the format is not a Format declared in configio.
func DecoderOf(format Format, in []byte) Decoder {
	r := bytes.NewReader(in)
	switch format {
	case FormatYAML:
		return NewYAMLDecoder(r)
	case FormatJSON:
		return NewJSONDecoder(r)
	default:
		panic(fmt.Sprintf("configio.DecoderOf: unexpected format: %q", format))
	}
}
