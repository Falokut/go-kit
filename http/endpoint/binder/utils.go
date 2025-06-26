package binder

import (
	"errors"
	"reflect"

	"github.com/Falokut/go-kit/utils/cases"
)

var (
	ErrDestNilPtr       = errors.New("dest must be a non-nil pointer")
	ErrDestNotStructPtr = errors.New("dest must point to a struct")
)

func getStructValue(dest any) (reflect.Value, error) {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return reflect.Value{}, ErrDestNilPtr
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, ErrDestNotStructPtr
	}
	return v, nil
}

func getFieldName(field reflect.StructField, tag string) string {
	// Try to get field name from tag; fallback to lowerCamelCase of field name
	tagValue, ok := field.Tag.Lookup(tag)
	if !ok || tagValue == "" {
		return cases.ToLowerCamelCase(field.Name)
	}
	return tagValue
}
