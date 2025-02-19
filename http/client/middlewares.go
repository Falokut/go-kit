package client

import (
	"context"

	"github.com/Falokut/go-kit/requestid"
)

func RequestId() Middleware {
	return func(next RoundTripper) RoundTripper {
		return RoundTripperFunc(func(ctx context.Context, request *Request) (*Response, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			request.Raw.Header.Set(requestid.RequestIdHeader, requestId)
			return next.RoundTrip(ctx, request)
		})
	}
}
