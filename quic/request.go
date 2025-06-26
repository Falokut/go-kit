package quic

import (
	"context"
	"fmt"
	"github.com/Falokut/go-kit/json"
	"time"
)

type contextKey string

const requestContextKey contextKey = "request"

func WithRequest(ctx context.Context, req *Request) context.Context {
	return context.WithValue(ctx, requestContextKey, req)
}

func RequestFromContext(ctx context.Context) (*Request, bool) {
	req, ok := ctx.Value(requestContextKey).(*Request)
	return req, ok
}

type Request struct {
	Address string        `json:"-"`
	Timeout time.Duration `json:"-"`
	Body    json.RawMessage
	Headers map[string]string
}

func DecodeRequest(stream Stream) (*Request, error) {
	var req Request
	decoder := json.NewDecoder(stream)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("decode request: %w", err)
	}
	return &req, nil
}
