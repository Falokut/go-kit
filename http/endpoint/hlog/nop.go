package hlog

import (
	"context"
	"net/http"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint"
)

func Noop() endpoint.LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return next(ctx, w, r)
		}
	}
}
