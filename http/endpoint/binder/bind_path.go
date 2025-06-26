package binder

import (
	"net/http"
	"reflect"

	"github.com/Falokut/go-kit/http/router"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

/**
 * BindPath extracts path parameters from the request and binds them to the provided destination struct.
 * It supports nested and embedded struct fields using the `path` struct tag.
 *
 * Parameters:
 *   - r: the incoming HTTP request
 *   - dest: a pointer to the destination struct to bind path parameters into
 *
 * Returns:
 *   - error: if any binding or conversion error occurs
 */
func BindPath(r *http.Request, dest any) error {
	v, err := getStructValue(dest)
	if err != nil {
		return err
	}

	params := router.ParamsFromRequest(r)
	return bindStructFields(params, v)
}

// bindStructFields binds values from path parameters into the struct value recursively.
func bindStructFields(params router.Params, v reflect.Value) error {
	t := v.Type()
	info := getStructInfo(t, PathTag)

	for _, fi := range info.fields {
		fieldValue := v.Field(fi.index)

		if !fieldValue.CanSet() {
			continue
		}

		// Handle anonymous (embedded) fields
		if fi.anonymous {
			if fi.isPtr && fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}

			subVal := fieldValue
			if fi.isPtr {
				subVal = fieldValue.Elem()
			}

			if subVal.Kind() == reflect.Struct {
				if err := bindStructFields(params, subVal); err != nil {
					return err
				}
			}
			continue
		}

		param := params.ByName(fi.fieldName)

		// If exact parameter is not found, try extracting nested struct values by prefix
		if param == "" {
			if (fi.isPtr && fieldValue.Type().Elem().Kind() == reflect.Struct) || fieldValue.Kind() == reflect.Struct {
				// Initialize pointer to struct if needed
				if fi.isPtr && fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}

				subVal := fieldValue
				if fi.isPtr {
					subVal = fieldValue.Elem()
				}

				if subVal.Kind() == reflect.Struct {
					// Create sub-param set with prefix stripped
					subParams := filterParamsWithPrefix(params, fi.fieldName+".")
					if len(subParams) > 0 {
						if err := bindStructFields(subParams, subVal); err != nil {
							return err
						}
						continue
					}
				}
			}
			continue
		}

		ok, err := unmarshalInputToField(param, fieldValue)
		if err != nil {
			return errors.WithMessagef(err, "unmarshal field %q", fi.fieldName)
		}
		if ok {
			continue
		}

		err = setWithProperType(fieldValue.Kind(), param, fieldValue)
		if err != nil {
			return errors.WithMessagef(err, "set field %q", fi.fieldName)
		}
	}

	return nil
}

/**
 * filterParamsWithPrefix returns a new router.Params containing only parameters
 * whose keys start with the given prefix. The prefix is stripped from the resulting keys.
 *
 * Parameters:
 *   - params: the full list of path parameters
 *   - prefix: the key prefix to filter by
 *
 * Returns:
 *   - router.Params: the filtered parameter list with adjusted keys
 */
func filterParamsWithPrefix(params router.Params, prefix string) router.Params {
	var filtered router.Params
	for _, p := range params {
		if len(p.Key) > len(prefix) && p.Key[:len(prefix)] == prefix {
			filtered = append(filtered, httprouter.Param{
				Key:   p.Key[len(prefix):],
				Value: p.Value,
			})
		}
	}
	return filtered
}
