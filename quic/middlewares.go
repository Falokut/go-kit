// nolint:ireturn
package quic

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/requestid"
	"github.com/pkg/errors"
)

const (
	DefaultMaxRequestBodySize = 64 * 1024 * 1024

	panicRecoverStackSize = 4 << 10
)

type Middleware func(next StreamHandler) StreamHandler

// QuicError represents an error that provide error code.
type QuicError interface {
	ErrorCode() uint64
}

func ChainMiddleware(handler StreamHandler, mws ...Middleware) StreamHandler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func Recovery() Middleware {
	return func(next StreamHandler) StreamHandler {
		return StreamHandlerFunc(func(ctx context.Context, stream Stream) (err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				e, ok := r.(error)
				if ok {
					err = e
				} else {
					err = fmt.Errorf("%v", r) // nolint:err113
				}

				stack := make([]byte, panicRecoverStackSize)
				length := runtime.Stack(stack, false)
				err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
			}()

			return next.HandleStream(ctx, stream)
		})
	}
}

// nolint:nonamedreturns
func ErrorHandler(logger log.Logger) Middleware {
	return func(next StreamHandler) StreamHandler {
		return StreamHandlerFunc(func(ctx context.Context, stream Stream) (err error) {
			err = next.HandleStream(ctx, stream)
			if err == nil {
				return nil
			}

			logFunc := log.LogLevelFuncForError(err, logger)
			logFunc(ctx, err)

			var quicErr QuicError
			if errors.As(err, &quicErr) {
				stream.CancelWrite(quicErr.ErrorCode())
				return err
			}

			stream.CancelWrite(InternalErrorCode)
			return err
		})
	}
}

func DecodeRequestMiddleware() Middleware {
	return func(next StreamHandler) StreamHandler {
		return StreamHandlerFunc(func(ctx context.Context, stream Stream) error {
			req, err := DecodeRequest(stream)
			if err != nil {
				return err
			}

			ctx = WithRequest(ctx, req)
			return next.HandleStream(ctx, stream)
		})
	}
}

// nolint:err113
func RequestId() Middleware {
	return func(next StreamHandler) StreamHandler {
		return StreamHandlerFunc(func(ctx context.Context, stream Stream) error {
			req, ok := RequestFromContext(ctx)
			if !ok {
				return fmt.Errorf("request not found in context")
			}

			requestId := req.Headers[requestid.RequestIdHeader]
			if requestId == "" {
				requestId = requestid.Next()
			}

			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.Any("requestId", requestId))

			return next.HandleStream(ctx, stream)
		})
	}
}

type streamWithMaxBytesReader struct {
	Stream
	limitedReader io.Reader
}

func (s streamWithMaxBytesReader) Read(n []byte) (int, error) {
	return s.limitedReader.Read(n)
}

func MaxRequestBodySize(maxBytes int64) Middleware {
	return func(next StreamHandler) StreamHandler {
		return StreamHandlerFunc(func(ctx context.Context, stream Stream) error {
			streamWithLimit := streamWithMaxBytesReader{
				Stream:        stream,
				limitedReader: io.LimitReader(stream, maxBytes),
			}
			return next.HandleStream(ctx, streamWithLimit)
		})
	}
}
