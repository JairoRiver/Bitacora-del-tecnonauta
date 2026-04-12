package db_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("bitacora_test"),
		tcpostgres.WithUsername("bitacora"),
		tcpostgres.WithPassword("bitacora"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer container.Terminate(ctx) //nolint:errcheck

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	if err := runMigrations(connStr); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}
	defer testPool.Close()

	os.Exit(m.Run())
}

// runMigrations aplica todas las migraciones usando golang-migrate con el driver pgx/v5.
func runMigrations(connStr string) error {
	_, b, _, _ := runtime.Caller(0)
	// b es la ruta absoluta de este archivo: .../internal/db/main_test.go
	// las migraciones están en .../migrations/
	migrationsDir := filepath.Join(filepath.Dir(b), "..", "..", "migrations")

	// golang-migrate pgx/v5 espera el scheme "pgx5://"
	pgx5URL := strings.Replace(connStr, "postgres://", "pgx5://", 1)

	m, err := migrate.New("file://"+migrationsDir, pgx5URL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// newQuerier crea un Querier respaldado por una transacción que se revierte al
// finalizar el test. Garantiza aislamiento entre tests sin necesidad de
// limpiar datos manualmente.
func newQuerier(t *testing.T) db.Querier {
	t.Helper()
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		tx.Rollback(context.Background()) //nolint:errcheck
	})
	return db.New(tx)
}
