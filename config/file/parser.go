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

// unmarshaledConfig is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type unmarshaledConfig struct {
	Request struct {
		Method      *string           `yaml:"method" json:"method"`
		URL         *string           `yaml:"url" json:"url"`
		QueryParams map[string]string `yaml:"queryParams" json:"queryParams"`
		Timeout     *string           `yaml:"timeout" json:"timeout"`
	} `yaml:"request" json:"request"`

	RunnerOptions struct {
		Requests      *int    `yaml:"requests" json:"requests"`
		Concurrency   *int    `yaml:"concurrency" json:"concurrency"`
		GlobalTimeout *string `yaml:"globalTimeout" json:"globalTimeout"`
	} `yaml:"runnerOptions" json:"runnerOptions"`
}
