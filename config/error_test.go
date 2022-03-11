package config_test

import (
	"errors"
	"testing"

	"github.com/benchttp/runner/config"
)

func TestInvalidConfigError_Error(t *testing.T) {
	e := config.InvalidConfigError{
		Errors: []error{
			errors.New("error 0"),
			errors.New("error 1\nwith new line"),
			errors.New("error 2"),
		},
	}

	exp := `
Invalid value(s) provided:
  - error 0
  - error 1
with new line
  - error 2`[1:]

	if got := e.Error(); got != exp {
		t.Errorf("unexpected message:\nexp %s\ngot %s", exp, got)
	}
}
