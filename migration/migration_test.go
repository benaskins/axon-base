package migration_test

import (
	"database/sql"
	"embed"
	"testing"

	"github.com/benaskins/axon-base/migration"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed testdata/migrations
var testMigrations embed.FS

const testDSN = "postgres://postgres@localhost:5432/workbench?sslmode=disable"

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skip("postgres unavailable:", err)
	}
	return db
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
		name,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("tableExists query: %v", err)
	}
	return exists
}

func TestMigrate_UpAndDown(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	// Ensure clean state.
	if _, err := db.Exec("DROP TABLE IF EXISTS test_items, schema_migrations"); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	if err := migration.Migrate(db, testMigrations, "testdata/migrations"); err != nil {
		t.Fatalf("Migrate up: %v", err)
	}
	if !tableExists(t, db, "test_items") {
		t.Fatal("expected test_items table to exist after up migration")
	}

	if err := migration.Down(db, testMigrations, "testdata/migrations"); err != nil {
		t.Fatalf("Migrate down: %v", err)
	}
	if tableExists(t, db, "test_items") {
		t.Fatal("expected test_items table to be dropped after down migration")
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	if _, err := db.Exec("DROP TABLE IF EXISTS test_items, schema_migrations"); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	if err := migration.Migrate(db, testMigrations, "testdata/migrations"); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	// Second call should be a no-op (ErrNoChange is swallowed).
	if err := migration.Migrate(db, testMigrations, "testdata/migrations"); err != nil {
		t.Fatalf("second Migrate (idempotent): %v", err)
	}

	// Cleanup.
	if err := migration.Down(db, testMigrations, "testdata/migrations"); err != nil {
		t.Fatalf("cleanup Down: %v", err)
	}
}

