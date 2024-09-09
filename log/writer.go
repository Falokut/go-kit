package log

import (
	"bytes"
	"context"
)

type LoggerWriter struct {
	logFunc func(ctx context.Context, msg any, fields ...Field)
}

func (l *LoggerWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	l.logFunc(context.Background(), string(p))
	return len(p), nil
}
