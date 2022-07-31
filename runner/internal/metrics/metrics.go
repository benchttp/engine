package metrics

import "reflect"

// Value is a concrete metric value, e.g. 120 or 3 * time.Second.
type Value interface{}

// Field references an Aggregate field.
// It exposes a method Type that returns its intrisic type.
// It can be used to retrieve a Metric from an Aggregate
// via Aggregate.MetricOf(field).
type Field string

const (
	ResponseTimeAvg     Field = "AVG"
	ResponseTimeMin     Field = "MIN"
	ResponseTimeMax     Field = "MAX"
	RequestFailCount    Field = "FAILURE_COUNT"
	RequestSuccessCount Field = "SUCCESS_COUNT"
	RequestCount        Field = "TOTAL_COUNT"
)

type fieldInfo struct {
	typ      Type
	metricOf func(Aggregate) Metric
}

var fieldRecord = map[Field]fieldInfo{
	ResponseTimeAvg:     {TypeDuration, metricGetter(ResponseTimeAvg)},
	ResponseTimeMin:     {TypeDuration, metricGetter(ResponseTimeMin)},
	ResponseTimeMax:     {TypeDuration, metricGetter(ResponseTimeMax)},
	RequestFailCount:    {TypeInt, metricGetter(RequestFailCount)},
	RequestSuccessCount: {TypeInt, metricGetter(RequestSuccessCount)},
	RequestCount:        {TypeInt, metricGetter(RequestCount)},
}

func metricGetter(field Field) func(Aggregate) Metric {
	getter := func(getValue func(Aggregate) Value) func(Aggregate) Metric {
		return func(agg Aggregate) Metric {
			return Metric{Field: field, Value: getValue(agg)}
		}
	}
	switch field {
	case ResponseTimeAvg:
		return getter(func(agg Aggregate) Value { return agg.Avg })
	case ResponseTimeMin:
		return getter(func(agg Aggregate) Value { return agg.Min })
	case ResponseTimeMax:
		return getter(func(agg Aggregate) Value { return agg.Max })
	case RequestFailCount:
		return getter(func(agg Aggregate) Value { return agg.FailureCount })
	case RequestSuccessCount:
		return getter(func(agg Aggregate) Value { return agg.SuccessCount })
	case RequestCount:
		return getter(func(agg Aggregate) Value { return agg.TotalCount })
	default:
		panic(badField(field))
	}
}

// Type represents the underlying type of a Value.
type Type uint8

const (
	lastGoReflectKind = reflect.UnsafePointer

	// TypeInt corresponds to type int.
	TypeInt = Type(reflect.Int)
	// TypeDuration corresponds to type time.Duration.
	TypeDuration = Type(lastGoReflectKind + iota)
)

// Type returns the field's intrisic type.
func (field Field) Type() Type {
	return retrieveFieldOrPanic(field).typ
}

// MetricOf returns the Metric for the given Source in Aggregate.
func (agg Aggregate) MetricOf(field Field) Metric {
	return retrieveFieldOrPanic(field).metricOf(agg)
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

// retrieveFieldInfoOrPanic retrieves the fieldInfo for the given field.
//
// It panics if the field is not defined in fieldRecord.
func retrieveFieldOrPanic(field Field) fieldInfo {
	fieldData, ok := fieldRecord[field]
	if !ok {
		panic(badField(field))
	}
	return fieldData
}

func badField(field Field) string {
	return "metrics: unknown field: " + string(field)
}
