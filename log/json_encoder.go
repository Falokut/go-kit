// nolint:exhaustive,dupl
package log

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Falokut/go-kit/json"
)

type JsonEncoder struct{}

func (j JsonEncoder) Encode(fields ...Field) ([]byte, error) {
	builder := bpool.Get().(*strings.Builder) // nolint:forcetypeassert
	defer bpool.Put(builder)
	builder.Reset()

	builder.WriteString("{")
	for i, field := range fields {
		if i > 0 {
			builder.WriteString(", ")
		}
		encodeField(field, builder, j)
	}
	builder.WriteString("}")

	return []byte(builder.String()), nil
}

func (JsonEncoder) writeFieldName(name string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf(`"%s": `, name))
}

func (JsonEncoder) encodeReflectTypeFallback(value any, builder *strings.Builder) {
	_ = json.NewEncoder(builder).Encode(value)
}

func (JsonEncoder) encodeFieldFallback(value Field, builder *strings.Builder) {
	_ = json.NewEncoder(builder).Encode(value.Interface)
}

func (j JsonEncoder) encodeDictionaryType(m any, builder *strings.Builder) {
	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Map {
		builder.WriteString("null")
		return
	}
	builder.WriteByte('{')
	iter := val.MapRange()

	i := 0
	for iter.Next() {
		if i > 0 {
			builder.WriteString(",")
		}

		key := iter.Key().Interface()
		k, ok := key.(string)
		if ok {
			builder.WriteByte('"')
			builder.WriteString(escapeString(k))
			builder.WriteString(`"`)
			builder.WriteString(`":`)

			value := iter.Value().Interface()
			encodeField(Any(k, value), builder, j)
		}
		i++
	}

	builder.WriteByte('}')
}

func (j JsonEncoder) encodeObjectType(value any, builder *strings.Builder) {
	val := reflect.ValueOf(value)
	builder.WriteString("{")
	for i := range val.NumField() {
		if i > 0 {
			builder.WriteString(", ")
		}
		encodeField(Any(val.Type().Field(i).Name, val.Field(i).Interface()), builder, j)
	}
	builder.WriteString("}")
}

func (j JsonEncoder) encodeArrObjectType(value any, builder *strings.Builder) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice {
		builder.WriteString("null")
		return
	}

	builder.WriteString("[")
	for i := range val.Len() {
		if i > 0 {
			builder.WriteString(", ")
		}
		j.encodeObjectType(val.Index(i).Interface(), builder)
	}
	builder.WriteString("]")
}
