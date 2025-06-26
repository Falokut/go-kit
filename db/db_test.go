package db_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Falokut/go-kit/db"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := db.Config{
		Host:     envOrDefault("PG_HOST", "127.0.0.1"),
		Port:     port,
		Database: envOrDefault("PG_DB", "test"),
		Username: envOrDefault("PG_USER", "test"),
		Password: envOrDefault("PG_PASS", "test"),
		Params: map[string]string{
			"target_session_attrs": "read-write",
		},
	}
	db, err := db.Open(t.Context(), cfg, db.WithCreateSchema(true))
	require.NoError(err)
	var time time.Time
	err = db.SelectRow(t.Context(), &time, "select now()")
	require.NoError(err)
}

func envOrDefault(name string, defValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}
	return defValue
}
