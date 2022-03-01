package configflags

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benchttp/runner/config"
)

// outValue implements flag.Value
type outValue struct {
	out *[]config.OutputStrategy
}

// String returns a string representation of outValue.out.
func (v outValue) String() string {
	return fmt.Sprint(v.out)
}

// Set reads input string as comma-separated values and appends the values
// to the key's values of the referenced header.
func (v outValue) Set(in string) error {
	values := strings.Split(in, ",")
	if len(values) < 1 {
		return errors.New(`expect comma-separated values`)
	}
	for _, value := range values {
		*v.out = append(*v.out, config.OutputStrategy(value))
	}
	return nil
}
