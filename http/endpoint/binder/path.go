package binder

import (
	"net/http"
	"reflect"

	"github.com/Falokut/go-kit/http/router"
	"github.com/Falokut/go-kit/utils/cases"
)

const pathTag = "path"

func bindPath(r *http.Request, dest any) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()
	params := router.ParamsFromRequest(r)

	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)
		fieldKind := fieldValue.Kind()
		if field.Anonymous {
			if fieldKind == reflect.Ptr {
				fieldValue = fieldValue.Elem()
			}
		}
		if !fieldValue.CanSet() {
			continue
		}

		fieldName := field.Name
		pathParamName, ok := t.Field(i).Tag.Lookup(pathTag)
		if !ok {
			pathParamName = cases.ToLowerCamelCase(fieldName)
		}
		pathParam := params.ByName(pathParamName)
		if pathParam == "" {
			continue
		}

		if ok, err := unmarshalInputToField(fieldKind, pathParam, fieldValue); ok {
			if err != nil {
				return err
			}
			continue
		}

		err := setWithProperType(fieldKind, pathParam, fieldValue)
		if err != nil {
			return err
		}
	}
	return nil
}
