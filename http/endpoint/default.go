package endpoint

import (
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint/binder"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/validator"
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...http.Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		ResponseWriterParam(),
		RequestParam(),
		RangeParam(),
		BearerTokenParam(),
	}
	middlewares := append(
		[]http.Middleware{
			MaxRequestBodySize(defaultMaxRequestBodySize),
			RequestId(),
			DefaultLog(logger),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)

	return NewWrapper(
		paramMappers,
		binder.NewRequestBinder(validator.Default),
		DefaultResponseMapper{},
		logger,
	).WithMiddlewares(middlewares...)
}
