package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.ConfigGlobal.
func JSON(in []byte) (runner.Config, error) {
	parser := JSONParser{}

	var repr Representation
	if err := parser.Parse(in, &repr); err != nil {
		return runner.Config{}, err
	}

	cfg, err := ParseRepresentation(repr)
	if err != nil {
		return runner.Config{}, err
	}

	return cfg.Override(runner.DefaultConfig()), nil
}
