package client

import (
	"context"
	"time"

	"github.com/Falokut/go-kit/quic"
)

type RequestBuilder struct {
	address string
	timeout time.Duration
	headers map[string]string
	body    []byte

	execute func(ctx context.Context, rb *RequestBuilder) (*quic.Response, error)
}

func NewRequestBuilder(
	address string,
	execute func(ctx context.Context, rb *RequestBuilder) (*quic.Response, error),
) *RequestBuilder {
	return &RequestBuilder{
		address: address,
		headers: make(map[string]string),
		execute: execute,
	}
}

func (rb *RequestBuilder) Address(addr string) *RequestBuilder {
	rb.address = addr
	return rb
}

func (rb *RequestBuilder) Timeout(d time.Duration) *RequestBuilder {
	rb.timeout = d
	return rb
}

func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

func (rb *RequestBuilder) Body(b []byte) *RequestBuilder {
	rb.body = b
	return rb
}

func (rb *RequestBuilder) Do(ctx context.Context) (*quic.Response, error) {
	return rb.execute(ctx, rb)
}
