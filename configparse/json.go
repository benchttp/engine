package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.Runner.
func JSON(in []byte) (runner.Runner, error) {
	parser := JSONParser{}
	repr := Representation{}
	if err := parser.Parse(in, &repr); err != nil {
		return runner.Runner{}, err
	}

	cfg := runner.DefaultRunner()
	if err := repr.ParseInto(&cfg); err != nil {
		return runner.Runner{}, err
	}

	return cfg, nil
}
