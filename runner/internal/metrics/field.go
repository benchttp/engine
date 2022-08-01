package metrics

import "reflect"

// Field is the name of an Aggregate field.
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

// fieldDefinition holds the necessary values to identify
// and manipulate a field.
// It consists of an intrinsic type and an accessor that binds
// the field to an Aggregate metric value.
type fieldDefinition struct {
	// typ is the intrisic type of the field.
	typ Type
	// boundValue is an accessor that binds a field
	// to the value it represents in an Aggregate.
	boundValue func(Aggregate) Value
}

// fieldDefinitions is a table of truth for fields.
// It maps all Field references to their intrinsic fieldDefinition.
var fieldDefinitions = map[Field]fieldDefinition{
	ResponseTimeAvg:     {TypeDuration, func(a Aggregate) Value { return a.Avg }},
	ResponseTimeMin:     {TypeDuration, func(a Aggregate) Value { return a.Min }},
	ResponseTimeMax:     {TypeDuration, func(a Aggregate) Value { return a.Max }},
	RequestFailCount:    {TypeInt, func(a Aggregate) Value { return a.FailureCount }},
	RequestSuccessCount: {TypeInt, func(a Aggregate) Value { return a.SuccessCount }},
	RequestCount:        {TypeInt, func(a Aggregate) Value { return a.TotalCount }},
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
// It panics if field is not defined in fieldDefinitions.
func (field Field) Type() Type {
	return field.mustRetrieve().typ
}

// valueIn returns the field's bound value in the given Aggregate.
// It panics if field is not defined in fieldDefinitions.
func (field Field) valueIn(agg Aggregate) Value {
	return field.mustRetrieve().boundValue(agg)
}

// mustRetrieve retrieves the fieldDefinition for the given field
// from fieldDefinitions, or panics if not found.
func (field Field) mustRetrieve() fieldDefinition {
	fieldData, ok := fieldDefinitions[field]
	if !ok {
		panic(badField(field))
	}
	return fieldData
}

func badField(field Field) string {
	return "metrics: unknown field: " + string(field)
}
