package metrics

import "reflect"

// Value is a concrete metric value, e.g. 120 or 3 * time.Second.
type Value interface{}

// Field represents the origin of a Metric.
// It exposes a method Type that returns the type of the metric.
// It can be used to reference a Metric in an Aggregate
// via Aggregate.MetricOf.
type Field string

const (
	ResponseTimeAvg     Field = "AVG"
	ResponseTimeMin     Field = "MIN"
	ResponseTimeMax     Field = "MAX"
	RequestFailCount    Field = "FAILURE_COUNT"
	RequestSuccessCount Field = "SUCCESS_COUNT"
	RequestCount        Field = "TOTAL_COUNT"
)

// Type represents the underlying type of a Value.
type Type uint8

const (
	lastGoReflectKind = reflect.UnsafePointer

	// TypeInt corresponds to type int.
	TypeInt = Type(reflect.Int)
	// TypeDuration corresponds to type time.Duration.
	TypeDuration = Type(lastGoReflectKind + iota)
)

// Type returns the underlying type of the metric field refers to.
func (field Field) Type() Type {
	switch field {
	case ResponseTimeAvg, ResponseTimeMin, ResponseTimeMax:
		return TypeDuration
	case RequestFailCount, RequestSuccessCount, RequestCount:
		return TypeInt
	}
	panic(badField(field))
}

// MetricOf returns the Metric for the given Field in Aggregate.
func (agg Aggregate) MetricOf(field Field) Metric {
	var v interface{}
	switch field {
	case ResponseTimeAvg:
		v = agg.Avg
	case ResponseTimeMin:
		v = agg.Min
	case ResponseTimeMax:
		v = agg.Max
	case RequestFailCount:
		v = agg.FailureCount
	case RequestSuccessCount:
		v = agg.SuccessCount
	case RequestCount:
		v = agg.TotalCount
	default:
		panic(badField(field))
	}
	return Metric{Field: field, Value: v}
}

// Metric represents an Aggregate metric. It links together a Field
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
// 	receiver := Metric{Value: 120}
// 	comparer := Metric{Value: 100}
// 	receiver.Compare(comparer) == SUP
//
// 	receiver := Metric{Value: 120 * time.Millisecond}
// 	comparer := Metric{Value: 100}
// 	receiver.Compare(comparer) // panics!
//
// 	receiver := Metric{Value: http.Header{}}
// 	comparer := Metric{Value: http.Header{}}
// 	receiver.Compare(comparer) // panics!
func (m Metric) Compare(to Metric) ComparisonResult {
	return compareMetrics(to, m)
}

func badField(field Field) string {
	return "metrics: unknown Field: " + string(field)
}
