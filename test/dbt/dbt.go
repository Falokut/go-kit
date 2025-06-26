// nolint:mnd
package dbt

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type TestDb struct {
	*db.Client
	must   must
	schema string
}

func New(t *test.Test, opts ...db.Option) *TestDb {
	dbConfig := Config(t)
	ctx, cancel := context.WithTimeout(t.T().Context(), 5*time.Second)
	defer cancel()
	opts = append([]db.Option{
		db.WithCreateSchema(true),
	}, opts...)
	cli, err := db.Open(ctx, dbConfig, opts...)
	t.Assert().NoError(err, "open test db cli, schema: %s", dbConfig.Schema)

	db := &TestDb{
		Client: cli,
		schema: dbConfig.Schema,
		must: must{
			db:     cli,
			assert: t.Assert(),
		},
	}
	t.T().Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func Config(t *test.Test) db.Config {
	cfg := t.Config()
	schema := fmt.Sprintf("test_%s", t.Id())
	dbConfig := db.Config{
		Host:        cfg.Optional().String("PG_HOST", "127.0.0.1"),
		Port:        cfg.Optional().Int("PG_PORT", 5432),
		Database:    cfg.Optional().String("PG_DB", "test"),
		Username:    cfg.Optional().String("PG_USER", "test"),
		Password:    cfg.Optional().String("PG_PASS", "test"),
		Schema:      schema,
		MaxOpenConn: runtime.NumCPU(),
	}
	return dbConfig
}

func (db *TestDb) DB() (*db.Client, error) {
	return db.Client, nil
}

func (db *TestDb) Must() must { // for test purposes
	return db.must
}

func (db *TestDb) Schema() string {
	return db.schema
}

func (db *TestDb) Close() error {
	_, err := db.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", db.schema))
	if err != nil {
		return errors.WithMessage(err, "drop schema")
	}
	err = db.Client.Close()
	return errors.WithMessage(err, "close db")
}

type must struct {
	db     *db.Client
	assert *require.Assertions
}

func (m must) ExecContext(ctx context.Context, query string, args ...any) sql.Result {
	res, err := m.db.ExecContext(ctx, query, args...)
	m.assert.NoError(err)
	return res
}

func (m must) SelectContext(ctx context.Context, resultPtr any, query string, args ...any) {
	err := m.db.SelectContext(ctx, resultPtr, query, args...)
	m.assert.NoError(err)
}

func (m must) SelectRow(ctx context.Context, resultPtr any, query string, args ...any) {
	err := m.db.GetContext(ctx, resultPtr, query, args...)
	m.assert.NoError(err)
}

func (m must) ExecNamed(ctx context.Context, query string, arg any) sql.Result {
	res, err := m.db.NamedExecContext(ctx, query, arg)
	m.assert.NoError(err)
	return res
}

func (m must) Count(ctx context.Context, query string, args ...any) int {
	value := 0
	m.SelectRow(ctx, &value, query, args...)
	return value
}
