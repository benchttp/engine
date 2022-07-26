package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.ConfigGlobal.
func JSON(in []byte) (runner.ConfigGlobal, error) {
	parser := jsonParser{}

	var uconf unmarshaledConfig
	if err := parser.parse(in, &uconf); err != nil {
		return runner.ConfigGlobal{}, err
	}

	pconf, err := newParsedConfig(uconf)
	if err != nil {
		return runner.ConfigGlobal{}, err
	}

	return runner.ConfigDefault().Override(pconf.ConfigGlobal, pconf.fields...), nil
}
