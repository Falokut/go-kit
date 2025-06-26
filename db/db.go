package db

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

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
	createSchema    bool
}

// nolint:gochecknoglobals
var (
	defaultMaxOpenConn = runtime.NumCPU() * 10
)

// nolint:mnd
func Open(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	dbCli, err := open(ctx, cfg.ConnStr(), opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "open db")
	}

	maxOpenConn := defaultMaxOpenConn
	if cfg.MaxOpenConn > 0 {
		maxOpenConn = cfg.MaxOpenConn
	}
	maxIdleConns := max(maxOpenConn/2, 2)
	dbCli.SetMaxOpenConns(maxOpenConn)
	dbCli.SetMaxIdleConns(maxIdleConns)
	dbCli.SetConnMaxIdleTime(90 * time.Second)
	dbCli.MapperFunc(cases.ToSnakeCase)

	isReadOnly, err := dbCli.IsReadOnly(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "check is read only connection")
	}

	isCustomSchema := cfg.Schema != "public" && cfg.Schema != ""
	if !isReadOnly && dbCli.createSchema && isCustomSchema {
		_, err = dbCli.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.Schema))
		if err != nil {
			return nil, errors.WithMessage(err, "exec create schema query")
		}
	}

	if isCustomSchema {
		err = dbCli.checkSchemaExistence(ctx, cfg.Schema)
		if err != nil {
			return nil, errors.WithMessage(err, "check schema existence")
		}
	}

	if !isReadOnly && dbCli.migrationRunner != nil {
		err = dbCli.migrationRunner.Run(ctx, dbCli.DB.DB)
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
		if p != nil { // rollback and repanic
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

func (db *Client) IsReadOnly(ctx context.Context) (bool, error) {
	var isReadOnly string
	err := db.QueryRowContext(ctx, "SHOW transaction_read_only").Scan(&isReadOnly)
	if err != nil {
		return false, err
	}
	return isReadOnly == "on", nil
}

func open(ctx context.Context, connStr string, opts ...Option) (*Client, error) {
	db := &Client{}

	for _, opt := range opts {
		opt(db)
	}

	cfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config")
	}
	cfg.Tracer = db.queryTracers

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

func (db *Client) checkSchemaExistence(ctx context.Context, schema string) error {
	query := `SELECT EXISTS (
		SELECT 1 FROM pg_namespace WHERE nspname = $1
	)`
	var exists bool
	err := db.QueryRowContext(ctx, query, schema).Scan(&exists)
	if err != nil {
		return errors.WithMessage(err, "check schema existence")
	}
	if !exists {
		return errors.Errorf("schema '%s' does not exist", schema)
	}
	return nil
}
