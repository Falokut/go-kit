package endpoint

import (
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint/binder"
	"github.com/Falokut/go-kit/http/endpoint/response"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/validator"
)

// LogMiddleware defines the type for logging middleware used in HTTP handlers.
type LogMiddleware http.Middleware

// DefaultWrapper creates a standard HTTP endpoint wrapper with a set of default middlewares.
//
// It includes request body size limit, request ID generation, logging, error handling, and panic recovery.
// Additional custom middlewares can be passed via restMiddlewares.
//
// This is the preferred entry point for quickly setting up endpoints with common functionality.
func DefaultWrapper(logger log.Logger, logMiddleware LogMiddleware, restMiddlewares ...http.Middleware) Wrapper {
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
			http.Middleware(logMiddleware),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)

	return NewWrapper(
		paramMappers,
		binder.NewRequestBinder(validator.Default),
		response.JsonMapper{},
		logger,
	).WithMiddlewares(middlewares...)
}
