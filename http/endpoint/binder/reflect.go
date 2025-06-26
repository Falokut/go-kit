package binder

import (
	"encoding"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

type BindUnmarshaler interface {
	// UnmarshalParam decodes and assigns a value from an form or query param.
	UnmarshalParam(param string) error
}

// bindMultipleUnmarshaler is used by binder to unmarshal multiple values from request at once to
// type implementing this interface. For example request could have multiple query fields `?a=1&a=2&b=test` in that case
// for `a` following slice `["1", "2"] will be passed to unmarshaller.
type bindMultipleUnmarshaler interface {
	UnmarshalParams(params []string) error
}

// nolint:cyclop,mnd
func derefPointer(field reflect.Value) reflect.Value {
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}
	return field
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	structField = derefPointer(structField)
	valueKind = structField.Kind()

	switch valueKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var bitSize int
		switch valueKind {
		case reflect.Int8:
			bitSize = 8
		case reflect.Int16:
			bitSize = 16
		case reflect.Int32:
			bitSize = 32
		case reflect.Int64, reflect.Int:
			bitSize = 64
		}
		return setIntField(val, bitSize, structField)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var bitSize int
		switch valueKind {
		case reflect.Uint8:
			bitSize = 8
		case reflect.Uint16:
			bitSize = 16
		case reflect.Uint32:
			bitSize = 32
		case reflect.Uint64, reflect.Uint:
			bitSize = 64
		}
		return setUintField(val, bitSize, structField)

	case reflect.Bool:
		return setBoolField(val, structField)

	case reflect.Float32, reflect.Float64:
		bitSize := 32
		if valueKind == reflect.Float64 {
			bitSize = 64
		}
		return setFloatField(val, bitSize, structField)

	case reflect.String:
		structField.SetString(val)
		return nil

	default:
		return errors.Errorf("unknown value type '%v'", valueKind)
	}
}

// nolint:wrapcheck
func unmarshalInputsToField(values []string, field reflect.Value) (bool, error) {
	field = derefPointer(field)
	fieldIValue := field.Addr().Interface()

	unmarshaler, ok := fieldIValue.(bindMultipleUnmarshaler)
	if ok {
		err := unmarshaler.UnmarshalParams(values)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// nolint:wrapcheck
func unmarshalInputToField(val string, field reflect.Value) (bool, error) {
	field = derefPointer(field)
	fieldIValue := field.Addr().Interface()

	switch unmarshaler := fieldIValue.(type) {
	case BindUnmarshaler:
		err := unmarshaler.UnmarshalParam(val)
		if err != nil {
			return false, err
		}
		return true, nil
	case encoding.TextUnmarshaler:
		err := unmarshaler.UnmarshalText([]byte(val))
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}

	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return errors.Wrap(err, "unmarshal field")
	}

	field.SetInt(intVal)
	return nil
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}

	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return errors.WithMessage(err, "parse uint")
	}

	field.SetUint(uintVal)
	return nil
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return errors.WithMessage(err, "parse bool")
	}

	field.SetBool(boolVal)
	return nil
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}

	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return errors.WithMessage(err, "parse float")
	}

	field.SetFloat(floatVal)
	return nil
}
