package dbx

import (
	"context"
	"database/sql"
	"github.com/Falokut/go-kit/log"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/Falokut/go-kit/db"
	"github.com/pkg/errors"
)

var (
	ErrClientIsNotInitialized = errors.New("client is not initialized")
)

const healthcheckTimeout = 500 * time.Millisecond

type Client struct {
	options []db.Option
	prevCfg *atomic.Value
	cli     *atomic.Pointer[db.Client]
	logger  log.Logger
}

func New(logger log.Logger, opts ...db.Option) *Client {
	prevCfg := &atomic.Value{}
	prevCfg.Store(db.Config{})
	return &Client{
		options: opts,
		prevCfg: prevCfg,
		cli:     &atomic.Pointer[db.Client]{},
		logger:  logger,
	}
}
func (c *Client) Upgrade(ctx context.Context, config db.Config) error {
	c.logger.Debug(ctx, "db client: received new config")

	if reflect.DeepEqual(c.prevCfg.Load(), config) {
		c.logger.Debug(ctx, "db client: configs are equal. skipping initialization")
		return nil
	}

	c.logger.Debug(ctx, "db client: initialization began")

	newCli, err := db.Open(ctx, config, c.options...)
	if err != nil {
		return errors.WithMessage(err, "open new client")
	}

	oldCli := c.cli.Swap(newCli)
	if oldCli != nil {
		_ = oldCli.Close()
	}

	c.logger.Debug(ctx, "db client: initialization done")

	c.prevCfg.Store(config)

	return nil
}

func (c *Client) DB() (*db.Client, error) {
	return c.db()
}

func (c *Client) Select(ctx context.Context, ptr any, query string, args ...any) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.Select(ctx, ptr, query, args...)
}

func (c *Client) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.SelectRow(ctx, ptr, query, args...)
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.Exec(ctx, query, args...)
}

func (c *Client) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.ExecNamed(ctx, query, arg)
}

func (c *Client) RunInTransaction(ctx context.Context, txFunc db.TxFunc, opts ...db.TxOption) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.RunInTransaction(ctx, txFunc, opts...)
}

func (c *Client) Close() error {
	c.logger.Debug(context.Background(), "db client: call close")
	c.prevCfg.Store(db.Config{})
	oldCli := c.cli.Swap(nil)
	if oldCli != nil {
		return oldCli.Close()
	}
	return nil
}

func (c *Client) Healthcheck(ctx context.Context) error {
	cli, err := c.db()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, healthcheckTimeout)
	defer cancel()
	_, err = cli.Exec(ctx, "SELECT 1")
	if err != nil {
		return errors.WithMessage(err, "exec")
	}
	return nil
}

func (c *Client) db() (*db.Client, error) {
	oldCli := c.cli.Load()
	if oldCli == nil {
		return nil, ErrClientIsNotInitialized
	}
	return oldCli, nil
}
