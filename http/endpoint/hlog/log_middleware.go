// Package hlog provides HTTP middleware for structured logging of requests and responses.
//
// It implements endpoint.LogMiddleware and integrates with the go-kit HTTP framework.
//
// This middleware logs HTTP method, URL, and optionally the request and response bodies,
// based on the provided configuration or default options. It supports logging body content
// for the following content types: "application/json" and "text/xml" by default.
//
// There are two ways to use the middleware:
//   - Use Log() for simple body logging toggle
//   - Use LogWithOptions() for fine-grained control via functional options
//
// Example usage:
//
//	mux := http.NewServeMux()
//	handler := hlog.Log(logger, true)(yourHandler)
//	mux.Handle("/endpoint", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		_ = handler(context.Background(), w, r)
//	}))
package hlog

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/endpoint/buffer"
	"github.com/Falokut/go-kit/log"
	"github.com/pkg/errors"
)

// nolint:gochecknoglobals
func defaultLogContentTypes() []string {
	return []string{
		"application/json",
		"text/xml",
	}
}

type logConfig struct {
	logBodyContentTypes []string
	logRequestBody      bool
	logResponseBody     bool
}

// Log returns a logging middleware that logs HTTP request and response metadata.
// If logBody is true, it also logs the request and response bodies
// for supported content types ("application/json", "text/xml").
func Log(logger log.Logger, logBody bool) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogContentTypes(),
		logRequestBody:      logBody,
		logResponseBody:     logBody,
	}
	return middleware(logger, cfg)
}

// LogWithOptions returns a logging middleware with fine-grained configuration.
// Use this when you want to selectively enable request/response body logging
// or customize the content types for which body logging is enabled.
func LogWithOptions(logger log.Logger, opts ...Option) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogContentTypes(),
		logRequestBody:      false,
		logResponseBody:     false,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return middleware(logger, cfg)
}

// middleware is the internal implementation of the logging middleware.
// It wraps the given handler, logs method, URL, status code, elapsed time,
// and optionally the request and response bodies based on the configuration.
func middleware(logger log.Logger, cfg *logConfig) endpoint.LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			buf := buffer.Acquire(w)
			defer buffer.Release(buf)

			now := time.Now()
			requestLogFields := []log.Field{
				log.String("method", r.Method),
				log.String("url", r.URL.String()),
			}
			requestContentType := r.Header.Get("Content-Type")
			if cfg.logRequestBody && matchContentType(requestContentType, cfg.logBodyContentTypes) {
				err := buf.ReadRequestBody(r.Body)
				if err != nil {
					return errors.WithMessage(err, "read request body for logging")
				}
				err = r.Body.Close()
				if err != nil {
					return errors.WithMessage(err, "close request reader")
				}
				r.Body = buffer.NewRequestBody(buf.RequestBody())

				requestLogFields = append(requestLogFields, log.ByteString("requestBody", buf.RequestBody()))
			}

			logger.Debug(ctx, "http handler: request", requestLogFields...)

			err := next(ctx, buf, r)

			responseLogFields := []log.Field{
				log.Int("statusCode", buf.StatusCode()),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			}
			responseContentType := buf.Header().Get("Content-Type")
			if cfg.logResponseBody && matchContentType(responseContentType, cfg.logBodyContentTypes) {
				responseLogFields = append(responseLogFields, log.ByteString("responseBody", buf.ResponseBody()))
			}

			logger.Debug(ctx, "http handler: response", responseLogFields...)

			return err
		}
	}
}

func matchContentType(contentType string, availableContentTypes []string) bool {
	return slices.ContainsFunc(availableContentTypes,
		func(content string) bool { return strings.HasPrefix(contentType, content) },
	)
}
