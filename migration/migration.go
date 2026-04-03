// Package migration provides a runner for goose SQL migrations
// using an embedded filesystem as the migration source.
package migration

import (
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"

	"github.com/pressly/goose/v3"
)

// Run runs all pending up migrations from the given filesystem and directory.
// The dir parameter is the path within fsys containing the SQL files
// (e.g., "migrations" for an embed.FS rooted at the module).
// It is safe to call multiple times; already-applied migrations are skipped.
func Run(db *sql.DB, fsys fs.FS, dir string) error {
	goose.SetBaseFS(fsys)
	goose.SetLogger(log.New(io.Discard, "", 0))

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration: set dialect: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("migration: up: %w", err)
	}

	slog.Info("database migrations complete")
	return nil
}

// Down rolls back all applied migrations.
func Down(db *sql.DB, fsys fs.FS, dir string) error {
	goose.SetBaseFS(fsys)
	goose.SetLogger(log.New(io.Discard, "", 0))

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration: set dialect: %w", err)
	}

	if err := goose.DownTo(db, dir, 0); err != nil {
		return fmt.Errorf("migration: down: %w", err)
	}

	return nil
}
