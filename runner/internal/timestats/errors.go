package timestats

import (
	"errors"
	"fmt"
)

// ErrEmptySlice is returned when working on an empty slice.
var ErrEmptySlice = errors.New("input slice is empty")

var ErrNotEnoughRecordsForDeciles = errors.New("not enough records to compute deciles (need at least 10)")

func ComputeError(name string) error {
	return fmt.Errorf("failed to compute stat: %s", name)
}
