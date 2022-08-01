package configio

import (
	"github.com/benchttp/engine/runner"
)

var _ Unmarshaler = (*DTO)(nil)

// DTO is a config adapter for IO communications, such as JSON or YAML.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
// It implements configio.Interface.
type DTO struct {
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

	Output struct {
		Silent   *bool   `yaml:"silent" json:"silent"`
		Template *string `yaml:"template" json:"template"`
	} `yaml:"output" json:"output"`
}

func (raw DTO) UnmarshalConfig(dst *runner.Config) error {
	return Unmarshal(raw, dst)
}
