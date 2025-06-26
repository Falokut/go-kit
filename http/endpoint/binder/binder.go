// nolint:wrapcheck
package binder

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/json"
	"github.com/pkg/errors"
)

type Validator interface {
	Validate(value any) (bool, map[string]string)
}

type RequestBinder struct {
	validator Validator
}

func NewRequestBinder(validator Validator) *RequestBinder {
	return &RequestBinder{
		validator: validator,
	}
}

/**
 * Bind takes the incoming HTTP request, extracts the necessary data based on the request method,
 * and binds it to the provided destination type. It handles query parameters, path parameters,
 * and request body binding. It also performs validation on the bound data using the validator.
 *
 * Parameters:
 *   - ctx: the context of the request
 *   - contentType: the content type of the request body
 *   - r: the incoming HTTP request
 *   - destType: the reflect.Type of the destination object
 *
 * Returns:
 *   - reflect.Value: the reflect.Value of the bound data
 *   - error: an error if any binding or validation fails
 */
func (b *RequestBinder) Bind(
	ctx context.Context,
	contentType string,
	r *http.Request,
	destType reflect.Type,
) (reflect.Value, error) {
	method := r.Method
	dest := reflect.New(destType)
	if method == http.MethodGet || method == http.MethodDelete || method == http.MethodHead {
		err := bindQuery(r, dest)
		if err != nil {
			return reflect.Value{},
				apierrors.NewBusinessError(http.StatusBadRequest, "invalid request query", err)
		}
	}
	err := BindPath(r, dest.Interface())
	if err != nil {
		return reflect.Value{},
			apierrors.NewBusinessError(http.StatusBadRequest, "invalid path params", err)
	}

	err = b.BindBody(contentType, r, dest)
	if err != nil {
		return reflect.Value{}, err
	}

	elem := dest.Elem()
	ok, details := b.validator.Validate(elem.Interface())
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

/**
 * BindBody binds the request body based on the content type.
 * It supports binding for form data, JSON, and XML content types.
 *
 * Parameters:
 * - ctype: the content type of the request
 * - r: the http.Request object
 * - dest: the reflect.Value destination to bind the request body to
 *
 * Returns an error if the binding process encounters any issues.
 */
func (b *RequestBinder) BindBody(
	contentType string,
	r *http.Request,
	dest reflect.Value,
) error {
	var err error
	switch {
	case strings.HasPrefix(contentType, MIMEApplicationForm),
		strings.HasPrefix(contentType, MIMEMultipartForm):
		err = bindForm(r, dest)
	case strings.HasPrefix(contentType, MIMEApplicationJSON):
		err = bindJson(r.Body, dest)
	case strings.HasPrefix(contentType, MIMEApplicationXML),
		strings.HasPrefix(contentType, MIMETextXML):
		err = bindXml(r.Body, dest)
	default:
		return nil
	}
	if err != nil {
		return apierrors.NewBusinessError(
			http.StatusBadRequest,
			"invalid request body",
			err,
		)
	}
	return nil
}

func formatDetails(details map[string]string) map[string]any {
	result := make(map[string]any, len(details))
	for k, v := range details {
		arr := []rune(k)
		arr[0] = unicode.ToLower(arr[0])
		result[string(arr)] = v
	}
	return result
}

func bindJson(reader io.Reader, dest reflect.Value) error {
	return json.NewDecoder(reader).Decode(dest.Interface())
}

func bindXml(reader io.Reader, dest reflect.Value) error {
	err := xml.NewDecoder(reader).Decode(dest.Interface())

	var ute *xml.UnsupportedTypeError
	var se *xml.SyntaxError
	switch {
	case errors.As(err, &ute):
		return apierrors.NewBusinessError(http.StatusBadRequest,
			fmt.Sprintf("Unsupported type error: type=%v, error=%v", ute.Type, ute.Error()), err)
	case errors.As(err, &se):
		return apierrors.NewBusinessError(http.StatusBadRequest,
			fmt.Sprintf("Syntax error: line=%v, error=%v", se.Line, se.Error()), err)
	case err != nil:
		return apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	default:
		return nil
	}
}

func bindForm(r *http.Request, dest reflect.Value) error {
	return BindData(r.Form, dest.Interface(), FormTag)
}

func bindQuery(r *http.Request, dest reflect.Value) error {
	return BindData(r.URL.Query(), dest.Interface(), QueryTag)
}
