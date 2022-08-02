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
	Request RequestDTO `yaml:"request" json:"request"`
	Runner  RunnerDTO  `yaml:"runner" json:"runner"`
	Output  OutputDTO  `yaml:"output" json:"output"`
}

type RequestDTO struct {
	Method      *string             `yaml:"method" json:"method"`
	URL         *string             `yaml:"url" json:"url"`
	QueryParams map[string]string   `yaml:"queryParams" json:"queryParams"`
	Header      map[string][]string `yaml:"header" json:"header"`
	Body        *struct {
		Type    string `yaml:"type" json:"type"`
		Content string `yaml:"content" json:"content"`
	} `yaml:"body" json:"body"`
}

type RunnerDTO struct {
	Requests       *int    `yaml:"requests" json:"requests"`
	Concurrency    *int    `yaml:"concurrency" json:"concurrency"`
	Interval       *string `yaml:"interval" json:"interval"`
	RequestTimeout *string `yaml:"requestTimeout" json:"requestTimeout"`
	GlobalTimeout  *string `yaml:"globalTimeout" json:"globalTimeout"`
}

func (in RunnerDTO) UnmarshalConfig(dst *runner.Config) error {
	fieldsSet := make([]string, 0, 5)
	setField := func(f string) {
		*dst = dst.WithFields(f)
	}

	defer func() {
		*dst = dst.WithFields(fieldsSet...)
	}()

	if requests := in.Requests; requests != nil {
		dst.Runner.Requests = *requests
		setField(runner.ConfigFieldRequests)
	}

	if concurrency := in.Concurrency; concurrency != nil {
		dst.Runner.Concurrency = *concurrency
		setField(runner.ConfigFieldConcurrency)
	}

	if interval := in.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return err
		}
		dst.Runner.Interval = parsedInterval
		setField(runner.ConfigFieldInterval)
	}

	if requestTimeout := in.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return err
		}
		dst.Runner.RequestTimeout = parsedTimeout
		setField(runner.ConfigFieldRequestTimeout)
	}

	if globalTimeout := in.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return err
		}
		dst.Runner.GlobalTimeout = parsedGlobalTimeout
		setField(runner.ConfigFieldGlobalTimeout)
	}

	return nil
}

type OutputDTO struct {
	Silent   *bool   `yaml:"silent" json:"silent"`
	Template *string `yaml:"template" json:"template"`
}

func (raw DTO) UnmarshalConfig(dst *runner.Config) error {
	return Unmarshal(raw, dst)
}
