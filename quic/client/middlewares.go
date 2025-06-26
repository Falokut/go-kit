package client

import (
	"context"

	"github.com/Falokut/go-kit/quic"
	"github.com/Falokut/go-kit/requestid"
)

func RequestId() Middleware {
	return func(next RoundTripper) RoundTripper {
		return RoundTripperFunc(func(ctx context.Context, req *quic.Request) (*quic.Response, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			if req.Headers == nil {
				req.Headers = make(map[string]string)
			}
			req.Headers[requestid.RequestIdHeader] = requestId

			return next.RoundTrip(ctx, req)
		})
	}
}
