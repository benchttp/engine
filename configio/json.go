package configio

import "github.com/benchttp/sdk/benchttp"

// JSON reads input bytes as JSON and unmarshals it into a benchttp.Runner.
func JSON(in []byte, dst *benchttp.Runner) error {
	repr := Representation{}
	if err := (JSONParser{}).Parse(in, &repr); err != nil {
		return err
	}
	return repr.ParseInto(dst)
}
