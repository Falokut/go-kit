package log

import "context"

type logKey struct{}

func ToContext(ctx context.Context, fields ...Field) context.Context {
	oldFields := ContextLogValues(ctx)
	newFields := make(Fields)
	for k, v := range oldFields {
		newFields[k] = v
	}
	for _, field := range fields {
		newFields[field.Name] = field.Value
	}

	return context.WithValue(ctx, logKey{}, newFields)
}

func ContextLogValues(ctx context.Context) Fields {
	fields, _ := ctx.Value(logKey{}).(Fields)
	return fields
}
