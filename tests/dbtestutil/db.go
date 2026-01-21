package dbtestutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const (
	testDatabaseURLDefault = "postgres://fintrack:fintrack@localhost:5432/fintrack_test"
)

func GetTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return testDatabaseURLDefault
}

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, GetTestDatabaseURL())
	require.NoError(t, err, "failed to connect to test database")

	require.NoError(t, pool.Ping(ctx), "failed to ping test database")

	return pool
}

func TeardownTestDB(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, "TRUNCATE TABLE transactions CASCADE")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "TRUNCATE TABLE categories CASCADE")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)
}

func CreateTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func AssertRowExists(t *testing.T, pool *pgxpool.Pool, query string, args ...interface{}) {
	ctx := CreateTestContext(t)
	var exists bool
	err := pool.QueryRow(ctx, query, args...).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "expected row to exist")
}

func AssertRowCount(t *testing.T, pool *pgxpool.Pool, expected int, query string, args ...interface{}) {
	ctx := CreateTestContext(t)
	var count int
	err := pool.QueryRow(ctx, query, args...).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, expected, count, fmt.Sprintf("expected %d rows, got %d", expected, count))
}

func RunInTransaction(t *testing.T, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) {
	ctx := CreateTestContext(t)
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			t.Errorf("transaction failed: %v", err)
			return
		}
		err = tx.Commit(ctx)
		require.NoError(t, err)
	}()

	err = fn(tx)
}