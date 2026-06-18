package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultMigrationsDir returns the postgres migrations directory.
func DefaultMigrationsDir() (string, error) {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir, nil
	}

	candidates := []string{
		"db/migrations/postgres",
		filepath.Join("..", "..", "db", "migrations", "postgres"),
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", errors.New("postgres migrations directory not found; set MIGRATIONS_DIR")
}

// Migrate applies pending SQL migrations against the given DSN.
func Migrate(dsn string) error {
	dir, err := DefaultMigrationsDir()
	if err != nil {
		return err
	}

	sourceURL := "file://" + filepath.ToSlash(dir)
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// NewPool connects to PostgreSQL and verifies connectivity.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	return pool, nil
}

// Open connects, migrates schema, and returns a ready connection pool.
func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	if err := Migrate(dsn); err != nil {
		return nil, err
	}
	return NewPool(ctx, dsn)
}
