package db

import (
	"github.com/Falokut/go-kit/db/migration"
	"github.com/Falokut/go-kit/log"
	"github.com/jackc/pgx/v5"
)

type Option func(cli *Client)

func WithMigrationRunner(migrationDir string, logger log.Logger) Option {
	return func(db *Client) {
		db.migrationRunner = migration.NewRunner(migration.DialectPostgreSQL, migrationDir, logger)
	}
}

func WithQueryTracer(tracers ...pgx.QueryTracer) Option {
	return func(db *Client) {
		db.queryTracers = append(db.queryTracers, tracers...)
	}
}

func WithCreateSchema(createSchema bool) Option {
	return func(db *Client) {
		db.createSchema = createSchema
	}
}
