package metrics

import (
	"errors"
	"reflect"
	"time"

	"github.com/benchttp/engine/internal/errorutil"
)

var ErrUnknownField = errors.New("metrics: unknown field")

// LegacyField is the name of an Aggregate field.
// It exposes a method Type that returns its intrisic type.
// It can be used to retrieve a Metric from an Aggregate
// via Aggregate.MetricOf(field).
type LegacyField string

const (
	ResponseTimeMin          LegacyField = "responseTimes.min"
	ResponseTimeMax          LegacyField = "responseTimes.max"
	ResponseTimeMean         LegacyField = "responseTimes.mean"
	EventTimeBodyReadMin     LegacyField = "eventTimes.bodyRead.min"
	EventTimeBodyReadMax     LegacyField = "eventTimes.bodyRead.max"
	EventTimeBodyReadMean    LegacyField = "eventTimes.bodyRead.mean"
	EventTimeFirstByteMin    LegacyField = "eventTimes.firstByte.min"
	EventTimeFirstByteMax    LegacyField = "eventTimes.firstByte.max"
	EventTimeFirstByteMean   LegacyField = "eventTimes.firstByte.mean"
	EventTimeConnectDoneMin  LegacyField = "eventTimes.connectDone.min"
	EventTimeConnectDoneMax  LegacyField = "eventTimes.connectDone.max"
	EventTimeConnectDoneMean LegacyField = "eventTimes.connectDone.mean"
	RequestFailureCount      LegacyField = "requests.failureCount"
	RequestSuccessCount      LegacyField = "requests.successCount"
	RequestCount             LegacyField = "requests.totalCount"
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
var fieldDefinitions = map[LegacyField]fieldDefinition{
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

	// TypeInvalid corresponds to an invalid type.
	TypeInvalid = Type(reflect.Invalid)
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
func (field LegacyField) Type() Type {
	return field.mustRetrieve().typ
}

// Validate returns ErrUnknownField if field is not a know Field, else nil.
func (field LegacyField) Validate() error {
	if !field.exists() {
		return errorutil.WithDetails(ErrUnknownField, field)
	}
	return nil
}

// func (field Field) IsCompatibleWith()

// valueIn returns the field's bound value in the given Aggregate.
// It panics if field is not defined in fieldDefinitions.
func (field LegacyField) valueIn(agg Aggregate) Value {
	return field.mustRetrieve().boundValue(agg)
}

func (field LegacyField) retrieve() (fieldDefinition, bool) {
	fieldData, exists := fieldDefinitions[field]
	return fieldData, exists
}

func (field LegacyField) exists() bool {
	_, exists := fieldDefinitions[field]
	return exists
}

// mustRetrieve retrieves the fieldDefinition for the given field
// from fieldDefinitions, or panics if not found.
func (field LegacyField) mustRetrieve() fieldDefinition {
	fieldData, exists := field.retrieve()
	if !exists {
		panic(badField(field))
	}
	return fieldData
}

func badField(field LegacyField) string {
	return "metrics: unknown field: " + string(field)
}

func (m Metric) Type() Type {
	switch m.Value.(type) {
	case int:
		return TypeInt
	case time.Duration:
		return TypeDuration
	default:
		return TypeInvalid
	}
}

type Field string

func (f Field) Type() Type {
	return Aggregate{}.MetricOf(f).Type()
}

func (f Field) Validate() error {
	if (Aggregate{}).MetricOf(f).Type() == TypeInvalid {
		return errorutil.WithDetails(ErrUnknownField, f)
	}
	return nil
}
