package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/utils/cases"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

type MigrationRunner interface {
	Run(ctx context.Context, db *sql.DB, gooseOpts ...goose.ProviderOption) error
}

type Client struct {
	*sqlx.DB
	queryTracers    tracers
	migrationRunner MigrationRunner
}

const defaultMaxOpenConn = 30

func NewDB(ctx context.Context, cfg config.Database, opts ...Option) (*Client, error) {
	cli := &Client{}

	for _, opt := range opts {
		opt(cli)
	}

	if cfg.Schema != "public" && cfg.Schema != "" {
		err := createSchema(ctx, cfg, cli.queryTracers)
		if err != nil {
			return nil, errors.WithMessage(err, "create schema")
		}
	}

	dbCli, err := open(ctx, cfg.ConnStr(), cli.queryTracers)
	if err != nil {
		return nil, errors.WithMessage(err, "open db")
	}

	maxOpenConn := defaultMaxOpenConn
	if cfg.MaxOpenConn > 0 {
		maxOpenConn = cfg.MaxOpenConn
	}
	maxIdleConns := maxOpenConn / 2
	if maxIdleConns < 2 {
		maxIdleConns = 2
	}
	dbCli.SetMaxOpenConns(maxOpenConn)
	dbCli.SetMaxIdleConns(maxIdleConns)
	dbCli.SetConnMaxIdleTime(90 * time.Second)
	dbCli.MapperFunc(cases.ToSnakeCase)

	if cli.migrationRunner != nil {
		err = cli.migrationRunner.Run(ctx, dbCli.DB.DB)
		if err != nil {
			return nil, errors.WithMessage(err, "migration run")
		}
	}

	return dbCli, nil
}

func (db *Client) RunInTransaction(ctx context.Context, txFunc TxFunc, opts ...TxOption) (err error) {
	options := &txOptions{}
	for _, opt := range opts {
		opt(options)
	}
	tx, err := db.BeginTxx(ctx, options.nativeOpts)
	if err != nil {
		return errors.WithMessage(err, "begin transaction")
	}
	defer func() {
		p := recover()
		if p != nil { //rollback and repanic
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = errors.WithMessage(err, rbErr.Error())
			}
			return
		}

		err = tx.Commit()
		if err != nil {
			err = errors.WithMessage(err, "commit tx")
		}
	}()

	return txFunc(ctx, &Tx{tx})
}

func (db *Client) Select(ctx context.Context, ptr any, query string, args ...any) error {
	return db.SelectContext(ctx, ptr, query, args...)
}

func (db *Client) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	return db.GetContext(ctx, ptr, query, args...)
}

func (db *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}

func (db *Client) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	return db.NamedExecContext(ctx, query, arg)
}

func createSchema(ctx context.Context, cfg config.Database, queryTracers tracers) error {
	dbCli, err := open(ctx, cfg.ConnStr(), queryTracers)
	if err != nil {
		return errors.WithMessage(err, "open db")
	}

	_, err = dbCli.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.Schema))
	if err != nil {
		return errors.WithMessage(err, "exec query")
	}

	err = dbCli.Close()
	if err != nil {
		return errors.WithMessage(err, "close db")
	}
	return nil
}

func open(ctx context.Context, connStr string, queryTracers tracers) (*Client, error) {
	db := &Client{}

	cfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config")
	}
	cfg.Tracer = queryTracers

	sqlDb := stdlib.OpenDB(*cfg)

	pgDb := sqlx.NewDb(sqlDb, "pgx")
	pgDb.MapperFunc(cases.ToSnakeCase)
	err = pgDb.PingContext(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "ping database")
	}

	db.DB = pgDb
	return db, nil
}
