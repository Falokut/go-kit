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
	migrationRunner MigrationRunner
}

const defaultMaxOpenConn = 30

func NewDB(ctx context.Context, cfg config.Database, opts ...Option) (*Client, error) {
	cli := &Client{}

	for _, opt := range opts {
		opt(cli)
	}

	if cfg.Schema != "public" && cfg.Schema != "" {
		err := createSchema(ctx, cfg)
		if err != nil {
			return nil, errors.WithMessage(err, "create schema")
		}
	}

	dbCli, err := open(ctx, cfg.ConnStr())
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

	if cli.migrationRunner != nil {
		err = cli.migrationRunner.Run(ctx, dbCli.DB.DB)
		if err != nil {
			return nil, errors.WithMessage(err, "migration run")
		}
	}

	return dbCli, nil
}

func open(ctx context.Context, connStr string) (*Client, error) {
	db := &Client{}

	cfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config")
	}

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

func createSchema(ctx context.Context, cfg config.Database) error {
	dbCli, err := open(ctx, cfg.ConnStr())
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
