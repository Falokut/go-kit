package app

import (
	"context"
	"os"

	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/log/file"
	"github.com/pkg/errors"
)

type Application struct {
	ctx    context.Context // nolint:containedctx
	cfg    *config.Config
	logger *log.Adapter

	cancel  context.CancelFunc
	runners []Runner
	closers []Closer
}

func New(opts ...Option) (*Application, error) {
	appConfig := DefaultConfig()
	for _, opt := range opts {
		opt(appConfig)
	}
	return NewFromConfig(*appConfig)
}

func NewFromConfig(appConfig Config) (*Application, error) {
	cfg, err := config.New(appConfig.ConfigOptions...)
	if err != nil {
		return nil, errors.WithMessage(err, "create config")
	}

	logger := getLogger(appConfig.LoggerConfigSupplier(cfg))
	ctx, cancel := context.WithCancel(context.Background())

	return &Application{
		logger: logger,
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		closers: []Closer{
			CloserFunc(func(context.Context) error {
				_ = logger.Close()
				return nil
			}),
		},
	}, nil
}

func (a *Application) Logger() *log.Adapter {
	return a.logger
}

func (a *Application) Context() context.Context {
	return a.ctx
}

func (a *Application) Config() *config.Config {
	return a.cfg
}

func (a *Application) AddRunners(runners ...Runner) {
	a.runners = append(a.runners, runners...)
}

func (a *Application) AddClosers(closers ...Closer) {
	a.closers = append(a.closers, closers...)
}

func (a *Application) Run() error {
	errChan := make(chan error)

	for i := range a.runners {
		go func(index int, runner Runner) {
			err := runner.Run(a.ctx)
			if err != nil {
				select {
				case errChan <- errors.WithMessagef(err, "start runner[%d] -> %T", index, runner):
				default:
					a.logger.Error(a.ctx, errors.WithMessagef(err, "start runner[%d] -> %T", index, runner))
				}
			}
		}(i, a.runners[i])
	}

	select {
	case err := <-errChan:
		return err
	case <-a.ctx.Done():
		return nil
	}
}

func (a *Application) Shutdown() {
	a.Close()
	a.cancel()
}

func (a *Application) Close() {
	for _, closer := range a.closers {
		err := closer.Close(a.ctx)
		if err != nil {
			a.logger.Error(a.ctx, errors.WithMessage(err, "run closer"))
		}
	}
}

func getLogger(cfg LogConfig) *log.Adapter {
	opts := make([]log.Option, 0)
	opts = append(opts, log.WithFieldsDeduplication(cfg.DeduplicateFields))

	switch cfg.EncoderType {
	case JsonEncoderType:
		opts = append(opts, log.WithEncoder(log.JsonEncoder{}))
	case PlainTextEncoderType:
		opts = append(opts, log.WithEncoder(log.PlainTextEncoder{}))
	}
	if cfg.FileOutput != nil {
		opts = append(opts, log.WithOutput(os.Stdout, file.NewFileOutput(*cfg.FileOutput)))
	}

	return log.New(opts...)
}
