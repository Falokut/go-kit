// nolint:exhaustive,dupl
package log

import (
	"fmt"
	"reflect"
	"strings"
)

type PlainTextEncoder struct{}

func (p PlainTextEncoder) Encode(fields ...Field) ([]byte, error) {
	builder := bpool.Get().(*strings.Builder) // nolint:forcetypeassert
	defer bpool.Put(builder)
	builder.Reset()

	for i, field := range fields {
		if i > 0 {
			builder.WriteString(" ")
		}
		p.encode(field, builder)
	}

	return []byte(builder.String()), nil
}

func (p PlainTextEncoder) encode(field Field, builder *strings.Builder) {
	encodeField(field, builder, p)
}

func (p PlainTextEncoder) writeFieldName(name string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf(`%s=`, name))
}

func (PlainTextEncoder) encodeReflectTypeFallback(value any, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("%v", value))
}

func (PlainTextEncoder) encodeFieldFallback(value Field, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("%v", value))
}

func (p PlainTextEncoder) encodeObjectType(value any, builder *strings.Builder) {
	val := reflect.ValueOf(value)

	builder.WriteString("{")
	for i := range val.NumField() {
		if i > 0 {
			builder.WriteString(", ")
		}
		fieldName := val.Type().Field(i).Name
		p.encode(Any(fieldName, val.Field(i).Interface()), builder)
	}
	builder.WriteString("}")
}

func (p PlainTextEncoder) encodeDictionaryType(m any, builder *strings.Builder) {
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
		value := iter.Value().Interface()
		k, ok := key.(string)
		if ok {
			builder.WriteByte('"')
			builder.WriteString(escapeString(k))
			builder.WriteString(`"`)
			builder.WriteString(`":`)
			encodeField(Any(k, value), builder, p)
		}
		i++
	}

	builder.WriteByte('}')
}

func (p PlainTextEncoder) encodeArrObjectType(value any, builder *strings.Builder) {
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
		item := val.Index(i).Interface()
		p.encodeArrObjectType(item, builder)
	}
	builder.WriteString("]")
}
