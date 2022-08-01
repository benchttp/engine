package tests

import (
	"errors"
	"fmt"
)

var ErrUnknownPredicate = errors.New("tests: unknown predicate")

func errWithDetails(err error, details interface{}) error {
	return fmt.Errorf("%w: %s", err, details)
}
