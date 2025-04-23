package log

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func encodePrimitiveSlice[T ~int | ~int64 | ~uint64 | ~float32 | ~float64 | ~bool](slice []T, builder *strings.Builder) {
	builder.WriteString("[")
	for i, v := range slice {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%v", v))
	}
	builder.WriteString("]")
}

func encodeStringSlice(slice []string, builder *strings.Builder) {
	builder.WriteString("[")
	for i, s := range slice {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf(`"%s"`, s))
	}
	builder.WriteString("]")
}

func escapeString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

type encoder interface {
	writeFieldName(name string, builder *strings.Builder)

	encodeObjectType(field any, builder *strings.Builder)
	encodeArrObjectType(field any, builder *strings.Builder)
	encodeDictionaryType(field any, builder *strings.Builder)

	encodeReflectTypeFallback(field any, builder *strings.Builder)
	encodeFieldFallback(field Field, builder *strings.Builder)
}

// nolint:forcetypeassert,cyclop,funlen,gosec
func encodeField(
	field Field,
	builder *strings.Builder,
	encoder encoder,
) {
	encoder.writeFieldName(field.Name, builder)
	switch field.Type {
	case StringType:
		builder.WriteString(fmt.Sprintf(`"%s"`, field.String))
	case IntType, UintType:
		builder.WriteString(strconv.Itoa(int(field.Int)))
	case Float32Type:
		builder.WriteString(fmt.Sprintf("%f", math.Float32frombits(uint32(field.Int))))
	case Float64Type:
		builder.WriteString(fmt.Sprintf("%f", math.Float64frombits(uint64(field.Int))))
	case BoolType:
		if field.Int == 1 {
			builder.WriteString("true")
			return
		}
		builder.WriteString("false")
	case TimeType:
		builder.WriteString(fmt.Sprintf(`"%s"`, field.Interface.(time.Time).Format(time.RFC3339)))
	case DurationType:
		builder.WriteString(fmt.Sprintf(`"%s"`, time.Duration(field.Int).String()))
	case ErrorType:
		builder.WriteString(fmt.Sprintf(`"%s"`, field.Interface.(error).Error()))
	case ObjectType:
		encoder.encodeObjectType(field.Interface, builder)
	case ReflectType:
		encodeReflectType(field.Interface, builder, encoder)
	case DictionaryType:
		encoder.encodeDictionaryType(field.Interface, builder)
	case ArrObjectType:
		encoder.encodeArrObjectType(field.Interface, builder)
	case ArrBoolType:
		encodePrimitiveSlice(field.Interface.([]bool), builder)
	case ArrIntType:
		encodePrimitiveSlice(field.Interface.([]int64), builder)
	case ArrUintType:
		encodePrimitiveSlice(field.Interface.([]uint64), builder)
	case ArrFloat32Type:
		encodePrimitiveSlice(field.Interface.([]float32), builder)
	case ArrFloat64Type:
		encodePrimitiveSlice(field.Interface.([]float64), builder)
	case ArrStringType:
		encodeStringSlice(field.Interface.([]string), builder)
	case ArrTimeType:
		times := field.Interface.([]time.Time)
		builder.WriteByte('[')
		for i, t := range times {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf(`"%s"`, t.Format(time.RFC3339)))
		}
		builder.WriteByte(']')
	case ArrDurationType:
		durations := field.Interface.([]time.Duration)
		builder.WriteByte('[')
		for i, d := range durations {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(d.String())
		}
		builder.WriteByte(']')
	case ArrReflectType:
		encodeArrReflectType(field.Interface, builder, encoder)
	case UnknownType:
		encoder.encodeFieldFallback(field, builder)
	}
}

func encodeReflectType(
	value any,
	builder *strings.Builder,
	encoder encoder,
) {
	switch v := value.(type) {
	case string:
		builder.WriteString(`"`)
		builder.WriteString(escapeString(v))
		builder.WriteString(`"`)
	case int, int8, int16, int32, int64:
		builder.WriteString(strconv.FormatInt(reflect.ValueOf(v).Int(), 10))
	case uint, uint8, uint16, uint32, uint64:
		builder.WriteString(strconv.FormatUint(reflect.ValueOf(v).Uint(), 10))
	case float32:
		builder.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		builder.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		builder.WriteString(strconv.FormatBool(v))
	case time.Time:
		builder.WriteString(`"`)
		builder.WriteString(v.Format(time.RFC3339))
		builder.WriteString(`"`)
	case time.Duration:
		builder.WriteString(`"`)
		builder.WriteString(v.String())
		builder.WriteString(`"`)
	case error:
		builder.WriteString(`"`)
		builder.WriteString(escapeString(v.Error()))
		builder.WriteString(`"`)
	case map[string]any:
		encoder.encodeDictionaryType(v, builder)
	default:
		encoder.encodeReflectTypeFallback(v, builder)
	}
}

func encodeArrReflectType(
	value any,
	builder *strings.Builder,
	encoder encoder,
) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		builder.WriteString("null")
		return
	}

	builder.WriteByte('[')
	for i := range val.Len() {
		if i > 0 {
			builder.WriteString(", ")
		}
		encodeReflectType(value, builder, encoder)
	}
	builder.WriteByte(']')
}
