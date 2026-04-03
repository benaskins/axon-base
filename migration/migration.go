// Package migration provides a runner for golang-migrate SQL migrations
// using an embedded filesystem as the migration source.
package migration

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func newMigrate(db *sql.DB, fsys fs.FS, dir string) (*migrate.Migrate, error) {
	src, err := iofs.New(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("migration: create source: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("migration: create driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("migration: new instance: %w", err)
	}
	return m, nil
}

// Migrate runs all pending up migrations from fsys/dir against db.
// It is safe to call multiple times; ErrNoChange is not treated as an error.
func Migrate(db *sql.DB, fsys fs.FS, dir string) error {
	m, err := newMigrate(db, fsys, dir)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration: up: %w", err)
	}
	return nil
}

// Down rolls back all applied migrations from fsys/dir against db.
func Down(db *sql.DB, fsys fs.FS, dir string) error {
	m, err := newMigrate(db, fsys, dir)
	if err != nil {
		return err
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration: down: %w", err)
	}
	return nil
}
