package log

import "context"

type Logger interface {
	Log(ctx context.Context, level Level, msg any, fields ...Field)
	Trace(ctx context.Context, msg any, fields ...Field)
	Debug(ctx context.Context, msg any, fields ...Field)
	Info(ctx context.Context, msg any, fields ...Field)
	Warn(ctx context.Context, msg any, fields ...Field)
	Warning(ctx context.Context, msg any, fields ...Field)
	Error(ctx context.Context, msg any, fields ...Field)
	Fatal(ctx context.Context, msg any, fields ...Field)
	Panic(ctx context.Context, msg any, fields ...Field)
}
