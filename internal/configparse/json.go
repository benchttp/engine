package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.ConfigGlobal.
func JSON(in []byte) (runner.Config, error) {
	parser := jsonParser{}

	var uconf unmarshaledConfig
	if err := parser.parse(in, &uconf); err != nil {
		return runner.Config{}, err
	}

	cfg, err := newParsedConfig(uconf)
	if err != nil {
		return runner.Config{}, err
	}

	return runner.DefaultConfig().Override(cfg), nil
}
