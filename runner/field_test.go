package runner_test

import (
	"testing"

	"github.com/drykit-go/testx"

	"github.com/benchttp/engine/runner"
)

func TestIsField(t *testing.T) {
	testx.Table(runner.IsConfigField).Cases([]testx.Case{
		{In: runner.ConfigFieldMethod, Exp: true},
		{In: runner.ConfigFieldURL, Exp: true},
		{In: runner.ConfigFieldHeader, Exp: true},
		{In: runner.ConfigFieldBody, Exp: true},
		{In: runner.ConfigFieldRequests, Exp: true},
		{In: runner.ConfigFieldConcurrency, Exp: true},
		{In: runner.ConfigFieldInterval, Exp: true},
		{In: runner.ConfigFieldRequestTimeout, Exp: true},
		{In: runner.ConfigFieldGlobalTimeout, Exp: true},
		{In: runner.ConfigFieldSilent, Exp: true},
		{In: runner.ConfigFieldTemplate, Exp: true},
		{In: "notafield", Exp: false},
	}).Run(t)
}
