package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.Runner.
func JSON(in []byte, dst *runner.Runner) error {
	repr := Representation{}
	if err := (JSONParser{}).Parse(in, &repr); err != nil {
		return err
	}
	return repr.ParseInto(dst)
}
