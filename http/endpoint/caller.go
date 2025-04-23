package endpoint

import (
	"context"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
)

// ResponseWriter is an interface that allows a response object
// to write itself directly to an http.ResponseWriter.
type ResponseWriter interface {
	Write(w http.ResponseWriter) error
}

// param represents a function parameter along with its index
// and a function for extracting its value from the HTTP request.
type param struct {
	index   int
	builder ParamBuilder
}

// Caller is responsible for invoking a user-defined handler function,
// preparing its arguments from an HTTP request, and writing the response.
type Caller struct {
	bodyExtractor RequestBinder
	bodyMapper    ResponseBodyMapper

	handler      reflect.Value
	paramsCount  int
	params       []param
	reqBodyIndex int
	reqBodyType  reflect.Type
}

// NewCaller creates a new Caller instance from the given handler function.
// It uses the provided RequestBinder to extract the request body (if applicable),
// and ParamMappers to resolve other arguments by their type.
//
// The function `f` must be a function. Parameters with known types will be resolved
// using the provided `paramMappers`, and one parameter may serve as the request body.
func NewCaller(
	f any,
	bodyExtractor RequestBinder,
	bodyMapper ResponseBodyMapper,
	paramMappers map[string]ParamMapper,
) (*Caller, error) {
	rt := reflect.TypeOf(f)
	if rt.Kind() != reflect.Func {
		return nil, errors.New("function expected")
	}

	paramsCount := rt.NumIn()
	reqBodyIndex := -1
	handler := reflect.ValueOf(f)
	var reqBodyType reflect.Type
	params := make([]param, 0)

	for i := range paramsCount {
		p := rt.In(i)
		paramType := p.String()
		mapper, ok := paramMappers[paramType]

		if !ok {
			// Only one request body is allowed
			if reqBodyIndex != -1 {
				return nil, errors.Errorf("param mapper not found for type %s", paramType)
			}
			reqBodyIndex = i
			reqBodyType = p
			continue
		}

		params = append(params, param{index: i, builder: mapper.Builder})
	}

	return &Caller{
		bodyExtractor: bodyExtractor,
		bodyMapper:    bodyMapper,
		handler:       handler,
		paramsCount:   paramsCount,
		params:        params,
		reqBodyIndex:  reqBodyIndex,
		reqBodyType:   reqBodyType,
	}, nil
}

// Handle prepares the arguments for the handler function from the HTTP request,
// calls the function, and writes the result to the http.ResponseWriter.
// If the result implements ResponseWriter, it is written directly; otherwise,
// it is passed to the configured ResponseBodyMapper.
func (h *Caller) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	args := make([]reflect.Value, h.paramsCount)

	if h.reqBodyIndex != -1 {
		contentType := r.Header.Get("Content-Type")
		value, err := h.bodyExtractor.Bind(ctx, contentType, r, h.reqBodyType)
		if err != nil {
			return err
		}
		args[h.reqBodyIndex] = value
	}

	for _, p := range h.params {
		value, err := p.builder(ctx, w, r)
		if err != nil {
			return err
		}
		args[p.index] = reflect.ValueOf(value)
	}

	returned := h.handler.Call(args)

	var result any
	var err error
	for i := range returned {
		v := returned[i]
		e, ok := v.Interface().(error)
		if ok && err == nil {
			err = e
		} else if result == nil && v.IsValid() {
			result = v.Interface()
		}
	}
	if err != nil {
		return err
	}

	writer, ok := result.(ResponseWriter)
	if ok {
		return writer.Write(w)
	}

	err = h.bodyMapper.Map(result, w)
	if err != nil {
		return err
	}

	return nil
}
