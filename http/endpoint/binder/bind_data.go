package binder

import (
	"reflect"

	"github.com/pkg/errors"
)

/**
 * BindData binds data from a map of string slices into the fields of a struct pointed to by dest.
 * The struct fields are matched based on struct tags or fallback field names converted to lowerCamelCase.
 *
 * Parameters:
 *   - valuesMap: a map of string keys to slices of strings representing input values, typically from form or query parameters.
 *   - dest: a pointer to a struct instance where the data should be bound.
 *   - tag: the struct tag key used to lookup field names in valuesMap (e.g., "form", "query").
 *
 * Returns:
 *   - error: an error if any binding or type conversion fails.
 *
 * Behavior:
 *   - Uses reflection to walk through the struct fields.
 *   - Recursively binds embedded (anonymous) structs and pointer-to-struct fields.
 *   - Supports slice types and initializes them element-wise.
 *   - Skips fields with tag `-` or unsettable fields.
 *   - Uses helper functions to parse and assign input values.
 *
 * Example:
 *
 *     var data MyStruct
 *     err := BindData(r.URL.Query(), &data, "query")
 *     if err != nil {
 *         // handle error
 *     }
 */
func BindData(valuesMap map[string][]string, dest any, tag string) error {
	return bindDataRecursive(valuesMap, dest, tag, "")
}

// bindDataRecursive is the internal implementation for BindData.
// It supports nested fields by accumulating a key prefix for structured input maps.
// nolint:gocognit,cyclop,funlen,nestif
func bindDataRecursive(valuesMap map[string][]string, dest any, tag, prefix string) error {
	v, err := getStructValue(dest)
	if err != nil {
		return err
	}

	t := v.Type()
	info := getStructInfo(t, tag)

	for _, fi := range info.fields {
		fieldValue := v.Field(fi.index)
		if !fieldValue.CanSet() {
			continue
		}

		key := prefix + fi.fieldName

		// Handle anonymous (embedded) structs
		if fi.anonymous {
			if fi.isPtr {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.Kind() == reflect.Struct {
				if err := bindDataRecursive(valuesMap, fieldValue.Addr().Interface(), tag, key+"."); err != nil {
					return errors.WithMessagef(err, "bind embedded struct field %q", key)
				}
			}
			continue
		}

		values := valuesMap[key]

		// Handle pointer to struct field
		if fi.isPtr && fieldValue.Kind() == reflect.Ptr &&
			fieldValue.Type().Elem().Kind() == reflect.Struct {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			err := bindDataRecursive(
				valuesMap,
				fieldValue.Interface(),
				tag,
				key+".",
			)
			if err != nil {
				return errors.WithMessagef(err, "bind pointer to struct field %q", key)
			}
			continue
		}

		// Handle nested struct field
		if len(values) == 0 {
			if fieldValue.Kind() == reflect.Struct {
				err := bindDataRecursive(
					valuesMap,
					fieldValue.Addr().Interface(),
					tag,
					key+".",
				)
				if err != nil {
					return errors.WithMessagef(err, "bind nested struct field %q", key)
				}
			}
			continue
		}

		// Try unmarshaling slice values
		ok, err := unmarshalInputsToField(values, fieldValue)
		if err != nil {
			return errors.WithMessagef(err, "unmarshal slice field %q", key)
		}
		if ok {
			continue
		}

		// Try unmarshaling single value
		ok, err = unmarshalInputToField(values[0], fieldValue)
		if err != nil {
			return errors.WithMessagef(err, "unmarshal field %q", key)
		}
		if ok {
			continue
		}

		// Dereference pointers
		fieldKind := fieldValue.Kind()
		if fieldKind == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem()
		}

		// Handle slice creation and assignment
		if fi.isSlice {
			numElems := len(values)
			slice := reflect.MakeSlice(fi.fieldType, numElems, numElems)
			for i := range values {
				err := setWithProperType(values[i], slice.Index(i))
				if err != nil {
					return errors.WithMessagef(err, "set slice element %d of field %q", i, key)
				}
			}
			fieldValue.Set(slice)
			continue
		}

		// Handle single value assignment
		err = setWithProperType(values[0], fieldValue)
		if err != nil {
			return errors.WithMessagef(err, "set field %q", key)
		}
	}

	return nil
}
