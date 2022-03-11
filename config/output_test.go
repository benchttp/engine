package config_test

import (
	"testing"

	"github.com/drykit-go/testx"

	"github.com/benchttp/runner/config"
)

func TestIsOutput(t *testing.T) {
	testx.Table(config.IsOutput).Cases([]testx.Case{
		{Lab: "valid lowercase", In: "benchttp", Exp: true},
		{Lab: "valid lowercase", In: "json", Exp: true},
		{Lab: "valid lowercase", In: "stdout", Exp: true},
		{Lab: "valid uppercase", In: "JSON", Exp: true},
		{Lab: "invalid", In: "notanoutput", Exp: false},
	}).Run(t)
}
