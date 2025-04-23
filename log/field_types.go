package log

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// A FieldType indicates which member of the Field union struct should be used
// and how it should be serialized.
type FieldType uint8

const (
	UnknownType FieldType = iota
	BoolType

	Float32Type
	Float64Type
	IntType
	UintType

	StringType
	TimeType
	DurationType
	ObjectType

	ErrorType

	ArrBoolType
	ArrFloat32Type
	ArrFloat64Type
	ArrIntType
	ArrUintType

	ArrStringType
	ArrTimeType
	ArrDurationType
	ArrObjectType

	ReflectType
	ArrReflectType
	DictionaryType
)

// for UTF-8 encoded string
func ByteString(name string, value []byte) Field {
	return String(name, string(value))
}

func Stringer(name string, value fmt.Stringer) Field {
	return String(name, value.String())
}

func String(name string, value string) Field {
	return Field{
		Name:   name,
		Type:   StringType,
		String: value,
	}
}

func Int(name string, value int) Field {
	return Int64(name, int64(value))
}

func Int8(name string, value int8) Field {
	return Int64(name, int64(value))
}

func Int16(name string, value int16) Field {
	return Int64(name, int64(value))
}

func Int32(name string, value int32) Field {
	return Int64(name, int64(value))
}

func Int64(name string, value int64) Field {
	return Field{
		Name: name,
		Type: IntType,
		Int:  value,
	}
}

func Uint(name string, value uint) Field {
	return Uint64(name, uint64(value))
}

func Uint8(name string, value uint8) Field {
	return Uint64(name, uint64(value))
}

func Uint16(name string, value uint16) Field {
	return Uint64(name, uint64(value))
}

func Uint32(name string, value uint32) Field {
	return Uint64(name, uint64(value))
}

func Uint64(name string, value uint64) Field {
	return Field{
		Name: name,
		Type: UintType,
		Int:  int64(value), // nolint:gosec
	}
}

func Float32(name string, value float32) Field {
	return Field{
		Name: name,
		Type: Float32Type,
		Int:  int64(math.Float32bits(value)),
	}
}

func Float64(name string, value float64) Field {
	return Field{
		Name: name,
		Type: Float64Type,
		Int:  int64(math.Float64bits(value)), // nolint:gosec
	}
}

func Bool(name string, value bool) Field {
	v := int64(0)
	if value {
		v = 1
	}
	return Field{
		Name: name,
		Type: BoolType,
		Int:  v,
	}
}

func Time(name string, value time.Time) Field {
	return Field{
		Name:      name,
		Type:      TimeType,
		Interface: value,
	}
}

func Duration(name string, value time.Duration) Field {
	return Field{
		Name: name,
		Type: DurationType,
		Int:  value.Nanoseconds(),
	}
}

func Object(name string, values any) Field {
	return Field{
		Name:      name,
		Type:      ObjectType,
		Interface: values,
	}
}

func Dictionary(name string, value map[string]any) Field {
	return Field{
		Name:      name,
		Type:      DictionaryType,
		Interface: value,
	}
}

func Error(err error) Field {
	return Field{
		Name:      FieldKeyLogError,
		Type:      ErrorType,
		Interface: err,
	}
}

func Bools(name string, values []bool) Field {
	return Field{
		Name:      name,
		Type:      ArrBoolType,
		Interface: values,
	}
}

func Floats32(name string, values []float32) Field {
	return Field{
		Name:      name,
		Type:      ArrFloat32Type,
		Interface: values,
	}
}

func Floats64(name string, values []float64) Field {
	return Field{
		Name:      name,
		Type:      ArrFloat64Type,
		Interface: values,
	}
}

func Ints(name string, values []int) Field {
	return Ints64(name, ints(values))
}

func Ints8(name string, values []int8) Field {
	return Ints64(name, ints(values))
}

func Ints16(name string, values []int16) Field {
	return Ints64(name, ints(values))
}

func Ints32(name string, values []int32) Field {
	return Ints64(name, ints(values))
}

func Ints64(name string, values []int64) Field {
	return Field{
		Name:      name,
		Type:      ArrIntType,
		Interface: values,
	}
}

func ints[T int | int8 | int16 | int32](values []T) []int64 {
	out := make([]int64, len(values))
	for i, v := range values {
		out[i] = int64(v)
	}
	return out
}

func Uints(name string, values []uint) Field {
	return Uints64(name, uints(values))
}

func Uints8(name string, values []uint8) Field {
	return Uints64(name, uints(values))
}

func Uints16(name string, values []uint16) Field {
	return Uints64(name, uints(values))
}

func Uints32(name string, values []uint32) Field {
	return Uints64(name, uints(values))
}

func Uints64(name string, values []uint64) Field {
	return Field{
		Name:      name,
		Type:      ArrUintType,
		Interface: values,
	}
}

func uints[T uint | uint8 | uint16 | uint32](values []T) []uint64 {
	out := make([]uint64, len(values))
	for i, v := range values {
		out[i] = uint64(v)
	}
	return out
}

func Strings(name string, values []string) Field {
	return Field{
		Name:      name,
		Type:      ArrStringType,
		Interface: values,
	}
}

func Times(name string, values []time.Time) Field {
	return Field{
		Name:      name,
		Type:      ArrTimeType,
		Interface: values,
	}
}

func Durations(name string, values []time.Duration) Field {
	return Field{
		Name:      name,
		Type:      ArrDurationType,
		Interface: values,
	}
}

func Objects(name string, values []any) Field {
	return Field{
		Name:      name,
		Type:      ArrObjectType,
		Interface: values,
	}
}

type anyFieldC[T any] func(string, T) Field

func (f anyFieldC[T]) Any(key string, val any) Field {
	v, _ := val.(T)
	return f(key, v)
}

// nolint:cyclop,gocyclo,funlen,inamedparam,exhaustive
func Any(name string, value any) Field {
	var c interface{ Any(string, any) Field }

	switch value.(type) {
	case bool:
		c = anyFieldC[bool](Bool)
	case []bool:
		c = anyFieldC[[]bool](Bools)

	case float64:
		c = anyFieldC[float64](Float64)
	case []float64:
		c = anyFieldC[[]float64](Floats64)

	case float32:
		c = anyFieldC[float32](Float32)
	case []float32:
		c = anyFieldC[[]float32](Floats32)

	case int:
		c = anyFieldC[int](Int)
	case []int:
		c = anyFieldC[[]int](Ints)

	case int64:
		c = anyFieldC[int64](Int64)
	case []int64:
		c = anyFieldC[[]int64](Ints64)

	case int32:
		c = anyFieldC[int32](Int32)
	case []int32:
		c = anyFieldC[[]int32](Ints32)

	case int16:
		c = anyFieldC[int16](Int16)
	case []int16:
		c = anyFieldC[[]int16](Ints16)

	case int8:
		c = anyFieldC[int8](Int8)
	case []int8:
		c = anyFieldC[[]int8](Ints8)

	case string:
		c = anyFieldC[string](String)
	case []string:
		c = anyFieldC[[]string](Strings)

	case uint:
		c = anyFieldC[uint](Uint)
	case []uint:
		c = anyFieldC[[]uint](Uints)

	case uint64:
		c = anyFieldC[uint64](Uint64)
	case []uint64:
		c = anyFieldC[[]uint64](Uints64)

	case uint32:
		c = anyFieldC[uint32](Uint32)
	case []uint32:
		c = anyFieldC[[]uint32](Uints32)

	case uint16:
		c = anyFieldC[uint16](Uint16)
	case []uint16:
		c = anyFieldC[[]uint16](Uints16)

	case uint8:
		c = anyFieldC[uint8](Uint8)
	case []uint8:
		c = anyFieldC[[]uint8](Uints8)

	case time.Time:
		c = anyFieldC[time.Time](Time)
	case []time.Time:
		c = anyFieldC[[]time.Time](Times)

	case time.Duration:
		c = anyFieldC[time.Duration](Duration)
	case []time.Duration:
		c = anyFieldC[[]time.Duration](Durations)

	case fmt.Stringer:
		c = anyFieldC[fmt.Stringer](Stringer)

	case error:
		c = anyFieldC[error](func(_ string, err error) Field {
			return Error(err)
		})
	default:
		reflectVal := reflect.ValueOf(value)
		switch reflectVal.Kind() {
		case reflect.Struct:
			c = anyFieldC[any](Object)
		case reflect.Slice, reflect.Array:
			switch reflectVal.Elem().Kind() {
			case reflect.Struct:
				c = anyFieldC[[]any](Objects)
			default:
				c = anyFieldC[any](func(name string, value any) Field {
					return Field{
						Name:      name,
						Type:      ArrReflectType,
						Interface: value,
					}
				})
			}
		case reflect.Map:
			c = anyFieldC[any](func(name string, value any) Field {
				return Field{
					Name:      name,
					Type:      DictionaryType,
					Interface: value,
				}
			})
		default:
			c = anyFieldC[any](func(name string, value any) Field {
				return Field{
					Name:      name,
					Type:      ReflectType,
					Interface: value,
				}
			})
		}
	}

	return c.Any(name, value)
}
