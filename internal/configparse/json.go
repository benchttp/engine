package configparse

import (
	"github.com/benchttp/runner/config"
)

// JSON reads input bytes as JSON and unmarshals it into a config.Global.
func JSON(in []byte) (config.Global, error) {
	parser := jsonParser{}

	var uconf unmarshaledConfig
	if err := parser.parse(in, &uconf); err != nil {
		return config.Global{}, err
	}

	pconf, err := newParsedConfig(uconf)
	if err != nil {
		return config.Global{}, err
	}

	return config.Default().Override(pconf.Global, pconf.fields...), nil
}
