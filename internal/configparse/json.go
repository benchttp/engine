package configparse

import (
	"github.com/benchttp/engine/runner"
)

// JSON reads input bytes as JSON and unmarshals it into a runner.ConfigGlobal.
//
// TODO: decorrelate from files logics: it has nothing to do with files.
// It could probably be moved to configio package, however this is not
// a trivial change if we want to keep the same JSON marshaling logics
// (such as error handling) without duplicating it.
func JSON(in []byte) (runner.Config, error) {
	parser := jsonParser{}

	var raw configFileRepr // inacurate: should be configio.DTO (no Extends field)
	if err := parser.parse(in, &raw); err != nil {
		return runner.Config{}, err
	}

	return raw.DTO.ConfigWithDefault()
}
