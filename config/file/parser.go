package file

import (
	"encoding/json"
	"errors"

	"gopkg.in/yaml.v3"
)

type extension string

const (
	extYML  extension = ".yml"
	extYAML extension = ".yaml"
	extJSON extension = ".json"
)

type configParser struct {
	parseFunc func(in []byte, dst interface{}) error
}

func (p configParser) parse(in []byte, dst interface{}) error {
	return p.parseFunc(in, dst)
}

func newParser(ext extension) (configParser, error) {
	switch ext {
	case extYML, extYAML:
		return configParser{parseFunc: yaml.Unmarshal}, nil
	case extJSON:
		return configParser{parseFunc: json.Unmarshal}, nil
	default:
		return configParser{}, errors.New("unsupported config format")
	}
}
