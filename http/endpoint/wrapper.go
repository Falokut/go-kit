// Package endpoint provides abstractions and utilities for building composable HTTP endpoints,
// including parameter mappers, request binding, response mapping, and middleware chaining.
package endpoint

import (
	"context"
	"net/http"
	"reflect"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint/binder"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/validator"
)

// HttpError represents an error that can be written directly to an HTTP response.
type HttpError interface {
	WriteError(w http.ResponseWriter) error
}

// RequestBinder defines an interface for binding an HTTP request body to a Go value.
type RequestBinder interface {
	Bind(ctx context.Context, contentType string, r *http.Request, reqBodyType reflect.Type) (reflect.Value, error)
}

// ResponseBodyMapper defines an interface for writing an HTTP response based on a function result.
type ResponseBodyMapper interface {
	Map(result any, w http.ResponseWriter) error
}

// ParamBuilder is a function that extracts a value from an HTTP request and wraps it for use as a parameter.
type ParamBuilder func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error)

// ParamMapper defines a named parameter builder used during request handling to resolve input parameters.
type ParamMapper struct {
	Type    string
	Builder ParamBuilder
}

// Wrapper provides the core structure for building HTTP handlers from endpoint functions.
//
// It supports dependency injection via parameter mappers, automatic request binding,
// response transformation, and customizable middleware chains.
type Wrapper struct {
	ParamMappers map[string]ParamMapper
	Binder       RequestBinder
	BodyMapper   ResponseBodyMapper
	Middlewares  []http2.Middleware
	Logger       log.Logger
}

// NewWrapper creates a new Wrapper instance with the provided parameter mappers, binder,
// response body mapper, and logger. Middleware can be added using WithMiddlewares.
func NewWrapper(
	paramMappers []ParamMapper,
	binder RequestBinder,
	bodyMapper ResponseBodyMapper,
	logger log.Logger,
) Wrapper {
	mappers := make(map[string]ParamMapper)
	for _, mapper := range paramMappers {
		mappers[mapper.Type] = mapper
	}
	return Wrapper{
		ParamMappers: mappers,
		Binder:       binder,
		BodyMapper:   bodyMapper,
		Logger:       logger,
	}
}

// Endpoint converts a regular function into an http.HandlerFunc,
// applying parameter mapping, binding, middleware, and response handling.
// Panics on invalid input signature.
func (m Wrapper) Endpoint(f any) http.HandlerFunc {
	caller, err := NewCaller(f, m.Binder, m.BodyMapper, m.ParamMappers)
	if err != nil {
		panic(err)
	}

	handler := caller.Handle
	for i := len(m.Middlewares) - 1; i >= 0; i-- {
		handler = m.Middlewares[i](handler)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r.Context(), w, r)
		if err != nil {
			m.Logger.Error(r.Context(), err)
		}
	}
}

// WithMiddlewares adds additional HTTP middlewares to the wrapper.
func (m Wrapper) WithMiddlewares(middlewares ...http2.Middleware) Wrapper {
	return Wrapper{
		ParamMappers: m.ParamMappers,
		Binder:       m.Binder,
		BodyMapper:   m.BodyMapper,
		Middlewares:  append(m.Middlewares, middlewares...),
		Logger:       m.Logger,
	}
}

// WithValidator returns a copy of the wrapper using a new validator adapter.
func (m Wrapper) WithValidator(validator validator.Adapter) Wrapper {
	return Wrapper{
		ParamMappers: m.ParamMappers,
		Binder:       binder.NewRequestBinder(validator),
		BodyMapper:   m.BodyMapper,
		Middlewares:  m.Middlewares,
		Logger:       m.Logger,
	}
}
