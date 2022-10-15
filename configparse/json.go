package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.ConfigGlobal.
func JSON(in []byte) (runner.Config, error) {
	parser := JSONParser{}
	repr := Representation{}
	if err := parser.Parse(in, &repr); err != nil {
		return runner.Config{}, err
	}

	cfg := runner.DefaultConfig()
	if err := repr.Unmarshal(&cfg); err != nil {
		return runner.Config{}, err
	}

	return cfg, nil
}
