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
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	if ok, err := unmarshalInputToField(valueKind, val, structField); ok {
		return err
	}

	// nolint:exhaustive
	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

// nolint:wrapcheck
func unmarshalInputsToField(valueKind reflect.Kind, values []string, field reflect.Value) (bool, error) {
	field = derefPointer(valueKind, field)
	fieldIValue := field.Addr().Interface()

	unmarshaler, ok := fieldIValue.(bindMultipleUnmarshaler)
	if ok {
		return true, unmarshaler.UnmarshalParams(values)
	}
	return false, nil
}

// nolint:wrapcheck
func unmarshalInputToField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	field = derefPointer(valueKind, field)
	fieldIValue := field.Addr().Interface()

	switch unmarshaler := fieldIValue.(type) {
	case BindUnmarshaler:
		return true, unmarshaler.UnmarshalParam(val)
	case encoding.TextUnmarshaler:
		return true, unmarshaler.UnmarshalText([]byte(val))
	}

	return false, nil
}

func derefPointer(valueKind reflect.Kind, field reflect.Value) reflect.Value {
	if valueKind == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}
	return field
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}

	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return errors.WithMessage(err, "parse int")
	}

	field.SetInt(intVal)
	return err
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
