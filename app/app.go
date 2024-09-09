// nolint:containedctx
package app

import (
	"context"

	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/log"
	"github.com/pkg/errors"
)

type RunnerFunc func(context.Context) error
type CloserFunc func(context.Context) error

type Application struct {
	logger  log.Logger
	context context.Context
	runners []RunnerFunc
	closers []CloserFunc

	close context.CancelFunc
}

func New() (*Application, error) {
	var logCfg config.Log
	err := config.Read(&logCfg)
	if err != nil {
		return nil, err
	}

	ctx, close := context.WithCancel(context.Background())
	level, err := log.ParseLogLevel(logCfg.LogLevel)
	if err != nil {
		panic(errors.WithMessage(err, "parse log level"))
	}
	logger, err := log.NewFromConfig(log.Config{
		Loglevel: level,
		Output: log.OutputConfig{
			Console:  logCfg.ConsoleOutput,
			Filepath: logCfg.Filepath,
		},
	})
	if err != nil {
		panic(errors.WithMessage(err, "logger from config"))
	}

	return &Application{
		logger:  logger,
		context: ctx,
		close:   close,
	}, nil
}

//nolint:ireturn
func (a *Application) GetLogger() log.Logger {
	return a.logger
}

func (a *Application) Context() context.Context {
	return a.context
}

func (a *Application) AddRunners(runners ...RunnerFunc) {
	a.runners = append(a.runners, runners...)
}

func (a *Application) AddClosers(closers ...CloserFunc) {
	a.closers = append(a.closers, closers...)
}

func (a *Application) Run() error {
	for _, runner := range a.runners {
		go func() {
			err := runner(a.context)
			if err != nil {
				a.logger.Fatal(a.context, errors.WithMessage(err, "run runner"))
			}
		}()
	}
	return nil
}

func (a *Application) Shutdown() {
	for _, closer := range a.closers {
		err := closer(a.context)
		if err != nil {
			a.logger.Error(a.context, errors.WithMessage(err, "run closer"))
		}
	}
	a.close()
}
