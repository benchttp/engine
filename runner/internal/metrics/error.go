package metrics

import (
	"errors"
	"fmt"
)

var ErrUnknownField = errors.New("metrics: unknown field")

func errWithDetails(err error, details interface{}) error {
	return fmt.Errorf("%w: %v", err, details)
}
