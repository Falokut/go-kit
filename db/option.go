package db

import (
	"github.com/Falokut/go-kit/db/migration"
	"github.com/Falokut/go-kit/log"
)

type Option func(cli *Client)

func WithMigrationRunner(migrationDir string, logger log.Logger) Option {
	return func(db *Client) {
		db.migrationRunner = migration.NewRunner(migration.DialectPostgreSQL, migrationDir, logger)
	}
}
