package configflags

import (
	"fmt"
	"strings"

	"github.com/benchttp/runner/config"
)

// bodyValue implements flag.Value
type bodyValue struct {
	body *config.Body
}

// String returns a string representation of the referenced body.
func (v bodyValue) String() string {
	return fmt.Sprint(v.body)
}

// Set reads input string in format "type:content" and sets
// the referenced body accordingly.
//
// Note: only type "raw" is supported at the moment.
func (v bodyValue) Set(raw string) error {
	errFormat := fmt.Errorf(`expect format "<type>:<content>", got "%s"`, raw)

	if raw == "" {
		return errFormat
	}

	split := strings.SplitN(raw, ":", 2)
	if len(split) != 2 {
		return errFormat
	}

	btype, bcontent := split[0], split[1]
	if bcontent == "" {
		return errFormat
	}

	switch btype {
	case "raw":
		*v.body = config.NewBody(btype, bcontent)
	// case "file":
	// 	// TODO
	default:
		return fmt.Errorf(`unsupported type: %s (only "raw" accepted)`, btype)
	}
	return nil
}
