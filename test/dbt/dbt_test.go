package dbt_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/dbt"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConnections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"Test1"},
		{"Test2"},
		{"Test3"},
		{"Test4"},
		{"Test5"},
	}

	for _, tt := range tests {
		t.Run(tt.name,
			func(t *testing.T) {
				t.Parallel()
				test, _ := test.New(t)
				db := dbt.New(test)

				_, err := db.DB()
				require.NoError(t, err)

				db.Must().ExecContext(t.Context(), "SELECT 1")
				time.Sleep(100 * time.Millisecond)
			},
		)
	}
}

func TestMaxConnections(t *testing.T) {
	t.Parallel()

	totalConnections := 2 * runtime.NumCPU()
	for i := range totalConnections {
		t.Run(fmt.Sprintf("Connection-%d", i), func(t *testing.T) {
			t.Parallel()
			test, _ := test.New(t)
			db := dbt.New(test)

			_, err := db.DB()
			require.NoError(t, err)
			db.Must().ExecContext(t.Context(), "SELECT 1")
		})
	}
}
