package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.ConfigGlobal.
func JSON(in []byte) (runner.Config, error) {
	parser := jsonParser{}

	var repr Representation
	if err := parser.parse(in, &repr); err != nil {
		return runner.Config{}, err
	}

	cfg, err := newParsedConfig(repr)
	if err != nil {
		return runner.Config{}, err
	}

	return runner.DefaultConfig().Override(cfg), nil
}
