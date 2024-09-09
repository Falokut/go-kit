package extractor

import (
	"context"
	"fmt"
	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"unicode"
)

type FormRequestExtractor struct {
	Validator Validator
}

func (ex FormRequestExtractor) Extract(ctx context.Context, r *http.Request, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)

	err := ex.bindForm(r.Form, instance.Interface(), "")
	if err != nil {
		return reflect.Value{}, apierrors.NewBusinessError(
			http.StatusBadRequest,
			"invalid request body",
			errors.New("invalid request body"),
		)
	}

	elem := instance.Elem()
	ok, details := ex.Validator.Validate(elem.Interface())
	if ok {
		return elem, nil
	}
	formattedDetails := formatDetails(details)
	return reflect.Value{}, apierrors.NewBusinessError(
		http.StatusBadRequest,
		"invalid request body",
		errors.Errorf("validation errors: %v", formattedDetails),
	).WithDetails(formattedDetails)
}

func (ex FormRequestExtractor) bindForm(formData url.Values, dest interface{}, parentName string) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	// Iterate over the struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		fieldName := field.Name
		if parentName != "" {
			fieldName = parentName + "." + fieldName
		}
		fieldName = lowerCaseFirstChar(fieldName)

		formValue := formData.Get(fieldName)

		if formValue == "" {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(formValue)
		case reflect.Int:
			intValue, err := strconv.Atoi(formValue)
			if err != nil {
				return fmt.Errorf("invalid value for field %s: %v", fieldName, err)
			}
			fieldValue.SetInt(int64(intValue))
		case reflect.Struct:
			err := ex.bindForm(formData, fieldValue.Interface(), fieldName)
			if err != nil {
				return fmt.Errorf("invalid value for field %s: %v", fieldName, err)
			}
		default:
			return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
		}
	}

	return nil
}

func lowerCaseFirstChar(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return ""
}
