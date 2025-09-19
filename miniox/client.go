package miniox

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/Falokut/go-kit/log"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

var (
	ErrClientIsNotInitialized = errors.New("client is not initialized")
	ErrMinioOffline           = errors.New("minio offline")
)

const hcDuration = 5 * time.Second

type Client struct {
	prevCfg *atomic.Value
	cli     *atomic.Pointer[minio.Client]
	logger  log.Logger
}

func New(logger log.Logger) *Client {
	prevCfg := &atomic.Value{}
	prevCfg.Store(Config{})
	return &Client{
		prevCfg: prevCfg,
		cli:     &atomic.Pointer[minio.Client]{},
		logger:  logger,
	}
}

func (c *Client) Upgrade(ctx context.Context, cfg Config, opts ...Option) error {
	c.logger.Debug(ctx, "minio client: received new config")

	if reflect.DeepEqual(c.prevCfg.Load(), cfg) {
		c.logger.Debug(ctx, "minio client: configs are equal. skipping initialization")
		return nil
	}

	c.logger.Debug(ctx, "minio client: initialization began")
	options := &minio.Options{}
	for _, opt := range opts {
		opt(options)
	}

	cli, err := minio.New(cfg.Endpoint, options)
	if err != nil {
		return errors.WithMessage(err, "new minio client")
	}

	_, err = cli.HealthCheck(hcDuration)
	if err != nil {
		return errors.WithMessage(err, "init minio healthcheck")
	}

	c.cli.Store(cli)
	c.logger.Debug(ctx, "minio client: initialization done")
	c.prevCfg.Store(cfg)

	return nil
}

func (c *Client) Healthcheck(_ context.Context) error {
	cli, err := c.Client()
	if err != nil {
		return errors.WithMessage(err, "get client")
	}
	if cli.IsOnline() {
		return nil
	}
	return ErrMinioOffline
}

func (c *Client) Client() (*minio.Client, error) {
	cli := c.cli.Load()
	if cli == nil {
		return nil, ErrClientIsNotInitialized
	}
	return cli, nil
}
