package configflags

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// headerValue implements flag.Value
type headerValue struct {
	header *http.Header
}

// String returns a string representation of the referenced header.
func (v headerValue) String() string {
	return fmt.Sprint(v.header)
}

// Set reads input string in format "key:value" and appends value
// to the key's values of the referenced header.
func (v headerValue) Set(raw string) error {
	keyval := strings.SplitN(raw, ":", 2)
	if len(keyval) != 2 {
		return errors.New(`expect format "<key>:<value>"`)
	}
	key, val := keyval[0], keyval[1]
	(*v.header)[key] = append((*v.header)[key], val)
	return nil
}
