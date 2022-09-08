package metrics

import (
	"strings"

	"github.com/benchttp/engine/runner/internal/reflectutil"
)

// Value is a concrete metric value, e.g. 120 or 3 * time.Second.
type Value interface{}

// Metric represents an Aggregate metric. It links together a field id
// and its Value from the Aggregate.
// It exposes a method Compare that compares its Value to another.
type Metric struct {
	Field Field
	Value Value
}

// Compare compares a Metric's value to another.
// It returns a ComparisonResult that indicates the relationship
// between the two values from the receiver's point of view.
//
// It panics if m and n are not of the same type,
// or if their type is not handled.
//
// Examples:
//
//	receiver := Metric{Value: 120}
//	comparer := Metric{Value: 100}
//	receiver.Compare(comparer) == SUP
//
//	receiver := Metric{Value: 120 * time.Millisecond}
//	comparer := Metric{Value: 100}
//	receiver.Compare(comparer) // panics!
//
//	receiver := Metric{Value: http.Header{}}
//	comparer := Metric{Value: http.Header{}}
//	receiver.Compare(comparer) // panics!
func (m Metric) Compare(to Metric) ComparisonResult {
	return compareMetrics(to, m)
}

// MetricOf returns the Metric for the given field id in Aggregate.
func (agg Aggregate) MetricOf(field Field) Metric {
	resolvedValue := pathResolver().ResolvePath(agg, string(field))
	if !resolvedValue.IsValid() {
		return Metric{}
	}
	return Metric{
		Field: field,
		Value: resolvedValue.Interface(),
	}
}

// typeOf returns a string representation of the metric's type
// represented by a field path.
func (agg Aggregate) typeOf(field Field) string {
	if typ := pathResolver().ResolvePathType(agg, string(field)); typ != nil {
		return typ.String()
	}
	return ""
}

var exposedPathPatterns = []string{
	"(?i)ResponseTimes.*",
	"(?i)StatusCodesDistribution.*",
	"(?i)RequestEventTimes.*",
	"(?i)Records.*",
	"(?i)RequestFailures.*",
	"(?i)Request(Failure|Success)?Count",
}

func pathResolver() reflectutil.PathResolver {
	return reflectutil.PathResolver{
		KeyMatcher:      strings.EqualFold,
		AllowedPatterns: exposedPathPatterns,
	}
}
