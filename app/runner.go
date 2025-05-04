package app

import "context"

type Runner interface {
	Run(ctx context.Context) error
}

type RunnerFunc func(context.Context) error

func (r RunnerFunc) Run(ctx context.Context) error {
	return r(ctx)
}
