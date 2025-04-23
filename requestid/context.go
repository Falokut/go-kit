package requestid

import "context"

type contextKey struct{}

// nolint:gochecknoglobals
var (
	contextKeyValue = contextKey{}
)

const (
	RequestIdHeader = "x-request-id"
)

func ToContext(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, contextKeyValue, requestId)
}

func FromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextKeyValue).(string)
	return value
}
