package metrics

import (
	"strings"

	"github.com/benchttp/engine/runner/internal/reflectutil"
)

// Get parses fieldID as a path from the Aggregate receiver
// and returns the resolved value.
func (m Aggregate) Get(fieldID string) Value {
	return reflectutil.ResolvePathFunc(m, fieldID, strings.EqualFold).Interface()
}
