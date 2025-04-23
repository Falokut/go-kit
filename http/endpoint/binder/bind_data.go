package binder

import (
	"reflect"

	"github.com/Falokut/go-kit/utils/cases"
)

// nolint:gocognit,funlen
func bindData(valuesMap map[string][]string, dest any) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)
		fieldKind := fieldValue.Kind()
		if field.Anonymous {
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue = fieldValue.Elem()
			}
		}
		if !fieldValue.CanSet() {
			continue
		}

		fieldName := cases.ToLowerCamelCase(field.Name)
		values := valuesMap[fieldName]
		if len(values) == 0 {
			continue
		}

		ok, err := unmarshalInputsToField(v.Kind(), values, fieldValue)
		if err != nil {
			return err
		}
		if ok {
			continue
		}

		ok, err = unmarshalInputToField(v.Kind(), values[0], fieldValue)
		if err != nil {
			return err
		}
		if ok {
			continue
		}

		if fieldKind == reflect.Pointer {
			fieldKind = fieldValue.Elem().Kind()
			fieldValue = fieldValue.Elem()
		}

		if field.Type.Kind() == reflect.Slice {
			sliceOf := fieldValue.Elem().Kind()
			numElems := len(values)
			slice := reflect.MakeSlice(field.Type, numElems, numElems)
			for j := range numElems {
				err := setWithProperType(sliceOf, values[j], slice.Index(j))
				if err != nil {
					return err
				}
			}
			fieldValue.Set(slice)
			continue
		}

		err = setWithProperType(fieldKind, values[0], fieldValue)
		if err != nil {
			return err
		}
	}

	return nil
}
