package configio

import (
	"bytes"
	"fmt"

	"github.com/benchttp/engine/benchttp"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

type Decoder interface {
	Decode(dst *benchttp.Runner) error
}

// DecoderOf returns the appropriate Decoder for the given Format.
// It panics if the format is not a Format declared in configio.
func DecoderOf(format Format, in []byte) Decoder {
	return decoderOf(format, in)
}

type decoder interface {
	Decoder
	decodeRepr(dst *representation) error
}

func decoderOf(format Format, in []byte) decoder {
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
