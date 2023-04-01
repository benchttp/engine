package errorutil

import (
	"fmt"
	"strings"
)

// WithDetails returns an error wrapping err, appended with a string
// representation of details separated by ": ".
//
// Example
//
//	var ErrNotFound = errors.New("not found")
//	err := WithDetails(ErrNotFound, "abc.jpg", "deleted yesterday")
//
//	errors.Is(err, ErrNotFound) == true
//	err.Error() == "not found: abc.jpg: deleted yesterday"
func WithDetails(base error, details ...interface{}) error {
	detailsStr := make([]string, len(details))
	for i := range details {
		detailsStr[i] = fmt.Sprint(details[i])
	}
	return fmt.Errorf("%w: %s", base, strings.Join(detailsStr, ": "))
}
