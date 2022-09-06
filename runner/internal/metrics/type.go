package metrics

import (
	"reflect"
	"time"
)

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

// Type returns a Metric's intrinsic type.
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
