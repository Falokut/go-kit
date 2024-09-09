package endpoint

import (
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint/extractor"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/validator"
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...http.Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		ResponseWriterParam(),
		RequestParam(),
	}
	middlewares := append(
		[]http.Middleware{
			MaxRequestBodySize(defaultMaxRequestBodySize),
			DefaultLog(logger),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)

	return NewWrapper(
		paramMappers,
		extractor.NewRequestBodyExtractor(validator.Default),
		JsonResponseMapper{},
		logger,
	).WithMiddlewares(middlewares...)
}
