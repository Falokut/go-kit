package endpoint

import (
	"context"
	"net/http"
	"reflect"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/log"
)

type HttpError interface {
	WriteError(w http.ResponseWriter) error
}

type RequestBinder interface {
	Bind(ctx context.Context, contentType string, r *http.Request, reqBodyType reflect.Type) (reflect.Value, error)
}

type ResponseBodyMapper interface {
	Map(result any, w http.ResponseWriter) error
}

type ParamBuilder func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error)

type ParamMapper struct {
	Type    string
	Builder ParamBuilder
}

type Wrapper struct {
	ParamMappers map[string]ParamMapper
	Binder       RequestBinder
	BodyMapper   ResponseBodyMapper
	Middlewares  []http2.Middleware
	Logger       log.Logger
}

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

func (m Wrapper) WithMiddlewares(middlewares ...http2.Middleware) Wrapper {
	return Wrapper{
		ParamMappers: m.ParamMappers,
		Binder:       m.Binder,
		BodyMapper:   m.BodyMapper,
		Middlewares:  append(m.Middlewares, middlewares...),
		Logger:       m.Logger,
	}

}
