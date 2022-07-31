package metrics

import "reflect"

// Value is a concrete metric value, e.g. 120 or 3 * time.Second.
type Value interface{}

// Source represents the origin of a Metric.
// It exposes a method Type that returns the type of the metric.
// It can be used to reference a Metric in an Aggregate
// via Aggregate.MetricOf.
type Source string

const (
	ResponseTimeAvg     Source = "AVG"
	ResponseTimeMin     Source = "MIN"
	ResponseTimeMax     Source = "MAX"
	RequestFailCount    Source = "FAILURE_COUNT"
	RequestSuccessCount Source = "SUCCESS_COUNT"
	RequestCount        Source = "TOTAL_COUNT"
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

// Type returns the underlying type of the metric src refers to.
func (src Source) Type() Type {
	switch src {
	case ResponseTimeAvg, ResponseTimeMin, ResponseTimeMax:
		return TypeDuration
	case RequestFailCount, RequestSuccessCount, RequestCount:
		return TypeInt
	}
	panic(badSource(src))
}

// MetricOf returns the Metric for the given Source in Aggregate.
func (agg Aggregate) MetricOf(src Source) Metric {
	var v interface{}
	switch src {
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
		panic(badSource(src))
	}
	return Metric{Source: src, Value: v}
}

// Metric represents an Aggregate metric. It links together a Value
// and its Source from the Aggregate.
// It exposes a method Compare that compares its Value to another.
type Metric struct {
	Source Source
	Value  Value
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

func badSource(src Source) string {
	return "metrics: unknown Source: " + string(src)
}
