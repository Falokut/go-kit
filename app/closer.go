package app

import "context"

type Closer interface {
	Close(ctx context.Context) error
}

type CloserFunc func(ctx context.Context) error

func (c CloserFunc) Close(ctx context.Context) error {
	return c(ctx)
}
