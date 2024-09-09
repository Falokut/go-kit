package extractor

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/pkg/errors"
)

type Validator interface {
	Validate(value any) (bool, map[string]string)
}

type RequestBodyExtractor struct {
	jsonExtractor JsonRequestExtractor
	formExtractor FormRequestExtractor
}

func NewRequestBodyExtractor(validator Validator) RequestBodyExtractor {
	return RequestBodyExtractor{
		jsonExtractor: JsonRequestExtractor{Validator: validator},
		formExtractor: FormRequestExtractor{Validator: validator},
	}
}

func (ex RequestBodyExtractor) Extract(ctx context.Context,
	ctype string,
	r *http.Request,
	reqBodyType reflect.Type,
) (reflect.Value, error) {
	switch {
	case strings.HasPrefix(ctype, MIMEApplicationForm), strings.HasPrefix(ctype, MIMEMultipartForm):
		return ex.formExtractor.Extract(ctx, r, reqBodyType)
	case strings.HasPrefix(ctype, MIMEApplicationXML), strings.HasPrefix(ctype, MIMETextXML):
		return reflect.Value{}, nil
	case strings.HasPrefix(ctype, MIMEApplicationJSON):
		return ex.jsonExtractor.Extract(ctx, r.Body, reqBodyType)
	}
	return reflect.Value{}, apierrors.New(
		http.StatusBadRequest,
		http.StatusBadRequest,
		"unsupported content-type",
		errors.New("unsupported content-type"),
	)
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
