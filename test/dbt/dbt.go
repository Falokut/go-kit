package dbt

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Falokut/go-kit/client/db"
	"github.com/Falokut/go-kit/config"
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
	schema := fmt.Sprintf("test_%s", t.Id()) //name must start from none digit
	dbConfig := config.Database{
		Host:        optionalEnvStr("PG_HOST", "127.0.0.1"),
		Port:        optionalEnvInt("PG_PORT", 5432),
		Database:    optionalEnvStr("PG_DB", "test"),
		Username:    optionalEnvStr("PG_USER", "test"),
		Password:    optionalEnvStr("PG_PASS", "test"),
		Schema:      schema,
		MaxOpenConn: runtime.NumCPU(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cli, err := db.NewDB(ctx, dbConfig, opts...)
	t.Assert().NoError(err, errors.WithMessagef(err, "open test db cli, schema: %s", schema))

	db := &TestDb{
		Client: cli,
		schema: schema,
		must: must{
			db:     cli,
			assert: t.Assert(),
		},
	}
	t.T().Cleanup(func() {
		err := db.Close()
		t.Assert().NoError(err)
	})
	return db
}

func (db *TestDb) DB() (*db.Client, error) {
	return db.Client, nil
}

func (db *TestDb) Must() must { //for test purposes
	return db.must
}

func (db *TestDb) Schema() string {
	return db.schema
}

func (db *TestDb) Close() error {
	_, err := db.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", db.schema))
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

func optionalEnvStr(name string, defaultVal string) string {
	opt := os.Getenv(name)
	if opt != "" {
		return opt
	}
	return defaultVal
}

func optionalEnvInt(name string, defaultVal int) int {
	opt := os.Getenv(name)
	if opt == "" {
		return defaultVal
	}
	optVal, err := strconv.ParseInt(opt, 10, 64)
	if err != nil {
		return defaultVal
	}
	return int(optVal)
}
