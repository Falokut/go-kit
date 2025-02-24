package endpoint

import (
	"context"
	"net/http"
	"strings"
	"time"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint/buffer"
	"github.com/Falokut/go-kit/log"

	"github.com/pkg/errors"
)

type LogMiddleware http2.Middleware

var defaultAvailableContentTypes = []string{
	"application/json",
	"text/xml",
}

func LogWithContentTypes(
	logger log.Logger,
	availableContentTypes []string,
	logRequestBody bool,
	logResponseBody bool,
) LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			buf := buffer.Acquire(w)
			defer buffer.Release(buf)

			now := time.Now()
			requestLogFields := []log.Field{
				log.Any("method", r.Method),
				log.Any("url", r.URL.String()),
			}
			requestContentType := r.Header.Get("Content-Type")
			if logRequestBody && matchContentType(requestContentType, availableContentTypes) {
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
				log.Any("statusCode", buf.StatusCode()),
				log.Any("elapsedTimeMs", time.Since(now).Milliseconds()),
			}
			responseContentType := buf.Header().Get("Content-Type")
			if logResponseBody && matchContentType(responseContentType, availableContentTypes) {
				responseLogFields = append(responseLogFields, log.ByteString("responseBody", buf.ResponseBody()))
			}

			logger.Debug(ctx, "http handler: response", responseLogFields...)
			return err
		}
	}
}

func Log(logger log.Logger, logRequestBody bool, logResponseBody bool) LogMiddleware {
	return LogWithContentTypes(logger, defaultAvailableContentTypes, logRequestBody, logResponseBody)
}

func Noop() LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return next(ctx, w, r)
		}
	}
}

func matchContentType(contentType string, availableContentTypes []string) bool {
	for _, content := range availableContentTypes {
		if strings.HasPrefix(contentType, content) {
			return true
		}
	}
	return false
}
