package binder

import (
	"reflect"

	"github.com/Falokut/go-kit/utils/cases"
)

func bindData(valuesMap map[string][]string, dest any) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
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
		if ok {
			if err != nil {
				return err
			}
			continue
		}

		ok, err = unmarshalInputToField(v.Kind(), values[0], fieldValue)
		if ok {
			if err != nil {
				return err
			}
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
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, values[j], slice.Index(j)); err != nil {
					return err
				}
			}
			fieldValue.Set(slice)
			continue
		}

		if err := setWithProperType(fieldKind, values[0], fieldValue); err != nil {
			return err
		}
	}

	return nil
}
