package metrics

import (
	"errors"

	"github.com/benchttp/engine/internal/errorutil"
)

// ErrUnknownField occurs when a Field is used with an invalid path.
var ErrUnknownField = errors.New("metrics: unknown field")

// Field is an id representing the path from an Aggregate to
// one of its metrics. It can be used to retrieve a Metric
// from an Aggregate via Aggregate.MetricOf(field).
// It exposes a method Type that returns the type of the
// targeted metric.
type Field string

// Type returns the intrinsic Type of the metric targeted
// by the Field receiver.
func (f Field) Type() string {
	return Aggregate{}.TypeOf(f)
}

// Validate returns an ErrUnknownField if it does not correspond
// to a valid path from an Aggregate.
func (f Field) Validate() error {
	if f.Type() == "" {
		return errorutil.WithDetails(ErrUnknownField, f)
	}
	return nil
}
