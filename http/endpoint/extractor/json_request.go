package extractor

import (
	"context"
	"io"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/json"
)

type JsonRequestExtractor struct {
	Validator Validator
}

func (j JsonRequestExtractor) Extract(ctx context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)
	err := json.NewDecoder(reader).Decode(instance.Interface())
	if err != nil {
		err = errors.WithMessage(err, "unmarshal json request body")
		return reflect.Value{}, apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	}

	elem := instance.Elem()

	ok, details := j.Validator.Validate(elem.Interface())
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
