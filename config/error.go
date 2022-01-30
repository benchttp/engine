package config

import (
	"errors"
)

// ErrInvalid is returned for any invalid Config value.
var ErrInvalid = errors.New("invalid config")
