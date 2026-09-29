//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"create test database pool: %v",
			err,
		)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf(
			"ping test database: %v",
			err,
		)
	}

	return db
}

func cleanDatabase(
	t *testing.T,
	db *pgxpool.Pool,
) {
	t.Helper()

	_, err := db.Exec(
		context.Background(),
		`
		TRUNCATE TABLE
			rates,
			telegram_subscriptions
		RESTART IDENTITY;
		`,
	)
	if err != nil {
		t.Fatalf(
			"clean test database: %v",
			err,
		)
	}
}
