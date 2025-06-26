package qlog

import (
	"bytes"
	"context"
	"io"

	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/quic"
)

type logConfig struct {
	logRequestBody  bool
	logResponseBody bool
}

func Log(logger log.Logger, logBody bool) quic.Middleware {
	cfg := &logConfig{
		logRequestBody:  logBody,
		logResponseBody: logBody,
	}
	return middleware(logger, cfg)
}

func middleware(logger log.Logger, cfg *logConfig) quic.Middleware {
	return func(next quic.StreamHandler) quic.StreamHandler {
		return quic.StreamHandlerFunc(func(ctx context.Context, stream quic.Stream) error {
			var reqBuf, respBuf bytes.Buffer

			wrappedStream := &wrappedStream{
				Stream: stream,
				reader: io.TeeReader(stream, &reqBuf),
				writer: &respBuf,
			}

			err := next.HandleStream(ctx, wrappedStream)
			if err != nil {
				return err
			}

			if cfg.logRequestBody {
				logger.Debug(ctx, "quic request body",
					log.ByteString("body", reqBuf.Bytes()))
			}

			if cfg.logResponseBody {
				logger.Debug(ctx, "quic response body",
					log.ByteString("body", respBuf.Bytes()))
			}

			if respBuf.Len() > 0 {
				_, err := stream.Write(respBuf.Bytes())
				if err != nil {
					return err
				}
			}

			return nil
		})
	}
}
