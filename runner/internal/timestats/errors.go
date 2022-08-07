package timestats

import (
	"errors"
	"fmt"
)

// ErrEmptySlice is returned when working on an empty slice.
var ErrEmptySlice = errors.New("input slice is empty")

func ComputeError(name string) error {
	return fmt.Errorf("failed to compute stat: %s", name)
}
