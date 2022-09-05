package metrics

import (
	"errors"
	"reflect"

	"github.com/benchttp/engine/internal/errorutil"
)

var ErrUnknownField = errors.New("metrics: unknown field")

// Field is the name of an Aggregate field.
// It exposes a method Type that returns its intrisic type.
// It can be used to retrieve a Metric from an Aggregate
// via Aggregate.MetricOf(field).
type Field string

const (
	ResponseTimeMin          Field = "responseTimes.min"
	ResponseTimeMax          Field = "responseTimes.max"
	ResponseTimeMean         Field = "responseTimes.mean"
	EventTimeBodyReadMin     Field = "eventTimes.bodyRead.min"
	EventTimeBodyReadMax     Field = "eventTimes.bodyRead.max"
	EventTimeBodyReadMean    Field = "eventTimes.bodyRead.mean"
	EventTimeFirstByteMin    Field = "eventTimes.firstByte.min"
	EventTimeFirstByteMax    Field = "eventTimes.firstByte.max"
	EventTimeFirstByteMean   Field = "eventTimes.firstByte.mean"
	EventTimeConnectDoneMin  Field = "eventTimes.connectDone.min"
	EventTimeConnectDoneMax  Field = "eventTimes.connectDone.max"
	EventTimeConnectDoneMean Field = "eventTimes.connectDone.mean"
	RequestFailureCount      Field = "requests.failureCount"
	RequestSuccessCount      Field = "requests.successCount"
	RequestCount             Field = "requests.totalCount"
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
	ResponseTimeMin:          {TypeDuration, func(a Aggregate) Value { return a.ResponseTimes.Min }},
	ResponseTimeMax:          {TypeDuration, func(a Aggregate) Value { return a.ResponseTimes.Max }},
	ResponseTimeMean:         {TypeDuration, func(a Aggregate) Value { return a.ResponseTimes.Mean }},
	EventTimeConnectDoneMin:  {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["ConnectDone"].Min }},
	EventTimeConnectDoneMax:  {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["ConnectDone"].Max }},
	EventTimeConnectDoneMean: {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["ConnectDone"].Mean }},
	EventTimeFirstByteMin:    {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["FirstByte"].Min }},
	EventTimeFirstByteMax:    {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["FirstByte"].Max }},
	EventTimeFirstByteMean:   {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["FirstByte"].Mean }},
	EventTimeBodyReadMin:     {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["BodyRead"].Min }},
	EventTimeBodyReadMax:     {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["BodyRead"].Max }},
	EventTimeBodyReadMean:    {TypeDuration, func(a Aggregate) Value { return a.RequestEventTimes["BodyRead"].Mean }},
	RequestFailureCount:      {TypeInt, func(a Aggregate) Value { return len(a.RequestFailures) }},
	RequestSuccessCount:      {TypeInt, func(a Aggregate) Value { return len(a.Records) - len(a.RequestFailures) }},
	RequestCount:             {TypeInt, func(a Aggregate) Value { return len(a.Records) }},
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

// String returns a human-readable representation of the field.
//
// Example:
//
//	TypeDuration.String() == "time.Duration"
//	Type(123).String() == "[unknown type]"
func (typ Type) String() string {
	switch typ {
	case TypeInt:
		return "int"
	case TypeDuration:
		return "time.Duration"
	default:
		return "[unknown type]"
	}
}

// Type returns the field's intrisic type.
// It panics if field is not defined in fieldDefinitions.
func (field Field) Type() Type {
	return field.mustRetrieve().typ
}

// Validate returns ErrUnknownField if field is not a know Field, else nil.
func (field Field) Validate() error {
	if !field.exists() {
		return errorutil.WithDetails(ErrUnknownField, field)
	}
	return nil
}

// func (field Field) IsCompatibleWith()

// valueIn returns the field's bound value in the given Aggregate.
// It panics if field is not defined in fieldDefinitions.
func (field Field) valueIn(agg Aggregate) Value {
	return field.mustRetrieve().boundValue(agg)
}

func (field Field) retrieve() (fieldDefinition, bool) {
	fieldData, exists := fieldDefinitions[field]
	return fieldData, exists
}

func (field Field) exists() bool {
	_, exists := fieldDefinitions[field]
	return exists
}

// mustRetrieve retrieves the fieldDefinition for the given field
// from fieldDefinitions, or panics if not found.
func (field Field) mustRetrieve() fieldDefinition {
	fieldData, exists := field.retrieve()
	if !exists {
		panic(badField(field))
	}
	return fieldData
}

func badField(field Field) string {
	return "metrics: unknown field: " + string(field)
}
