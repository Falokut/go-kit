package log

import (
	"context"
)

type logKey struct{}

func ToContext(ctx context.Context, fields ...Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	fields = append(fields, ContextLogValues(ctx)...)
	return context.WithValue(ctx, logKey{}, fields)
}

func ContextLogValues(ctx context.Context) []Field {
	fields, ok := ctx.Value(logKey{}).(*[]Field)
	if !ok {
		return nil
	}
	return *fields
}
