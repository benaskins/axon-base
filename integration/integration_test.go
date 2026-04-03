// Package integration_test exercises all axon-base packages together against
// a real Postgres instance. Each sub-test builds on a shared schema created
// with the migration runner. No mocks are used.
package integration_test

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"testing"

	axonerrors "github.com/benaskins/axon-base/errors"
	"github.com/benaskins/axon-base/migration"
	"github.com/benaskins/axon-base/pool"
	"github.com/benaskins/axon-base/repository"
	"github.com/benaskins/axon-base/scan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed testdata/migrations
var testMigrations embed.FS

const testDSN = "postgres://postgres@localhost:5432/workbench?sslmode=disable"

// integItem is the entity used across all integration sub-tests.
type integItem struct {
	ID   string
	Name string
}

// integItemRepo is a concrete Repository[integItem] backed by a pgxpool.Pool.
// It demonstrates the axon-base repository pattern with explicit SQL.
type integItemRepo struct {
	db *pgxpool.Pool
}

var _ repository.Repository[integItem] = (*integItemRepo)(nil)

func (r *integItemRepo) Create(ctx context.Context, item integItem) (integItem, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO integ_items (id, name) VALUES ($1, $2) RETURNING id, name`,
		item.ID, item.Name,
	)
	var out integItem
	if err := row.Scan(&out.ID, &out.Name); err != nil {
		return integItem{}, fmt.Errorf("integItemRepo.Create: %w", err)
	}
	return out, nil
}

func (r *integItemRepo) Get(ctx context.Context, id string) (integItem, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name FROM integ_items WHERE id = $1`, id,
	)
	var item integItem
	if err := row.Scan(&item.ID, &item.Name); err != nil {
		return integItem{}, fmt.Errorf("integItemRepo.Get: %w", err)
	}
	return item, nil
}

func (r *integItemRepo) Update(ctx context.Context, item integItem) (integItem, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE integ_items SET name = $2 WHERE id = $1 RETURNING id, name`,
		item.ID, item.Name,
	)
	var out integItem
	if err := row.Scan(&out.ID, &out.Name); err != nil {
		return integItem{}, fmt.Errorf("integItemRepo.Update: %w", err)
	}
	return out, nil
}

func (r *integItemRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM integ_items WHERE id = $1`, id)
	return axonerrors.WrapError("integItemRepo.Delete", err)
}

func (r *integItemRepo) List(ctx context.Context) ([]integItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM integ_items ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("integItemRepo.List: %w", err)
	}
	return scan.Rows(rows, func(i *integItem) []any {
		return []any{&i.ID, &i.Name}
	})
}

// openDB returns a *sql.DB for the migration runner, skipping if Postgres is unreachable.
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

// openRawPool opens a pgxpool.Pool directly, skipping if Postgres is unreachable.
func openRawPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	p, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	if err := p.Ping(ctx); err != nil {
		p.Close()
		t.Skip("postgres unavailable:", err)
	}
	return p
}

// openPool opens a pool.Pool wrapper, skipping if Postgres is unreachable.
func openPool(t *testing.T) *pool.Pool {
	t.Helper()
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	if !p.Healthy(ctx) {
		p.Close()
		t.Skip("postgres unavailable")
	}
	return p
}

// TestIntegration is the top-level integration suite. It migrates up before
// running sub-tests and migrates down afterward for a clean exit.
func TestIntegration(t *testing.T) {
	db := openDB(t)

	// Ensure clean state, then migrate up.
	if _, err := db.Exec(`DROP TABLE IF EXISTS integ_items, schema_migrations`); err != nil {
		db.Close()
		t.Fatalf("pre-test cleanup: %v", err)
	}
	if err := migration.Migrate(db, testMigrations, "testdata/migrations"); err != nil {
		db.Close()
		t.Fatalf("migrate up: %v", err)
	}
	t.Cleanup(func() {
		defer db.Close()
		if err := migration.Down(db, testMigrations, "testdata/migrations"); err != nil {
			t.Errorf("migrate down: %v", err)
		}
	})

	t.Run("Pool", testPool)
	t.Run("Transaction", testTransaction)
	t.Run("Repository", testRepository)
	t.Run("Scan", testScan)
	t.Run("Errors", testErrors)
}

// testPool verifies pool health, metrics, and that Close makes the pool unhealthy.
func testPool(t *testing.T) {
	ctx := context.Background()
	p := openPool(t)

	if !p.Healthy(ctx) {
		t.Fatal("expected pool to be healthy")
	}

	m := p.Metrics()
	if m.Max == 0 {
		t.Fatal("expected Max > 0")
	}

	data, err := m.HealthJSON()
	if err != nil {
		t.Fatalf("HealthJSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty health JSON")
	}

	p.Close()
	if p.Healthy(ctx) {
		t.Fatal("expected pool to be unhealthy after Close")
	}
}

// testTransaction exercises WithTransaction: one commit path and one rollback path.
func testTransaction(t *testing.T) {
	ctx := context.Background()
	p := openPool(t)
	defer p.Close()

	raw := openRawPool(t)
	defer raw.Close()

	// Commit path: insert inside transaction, verify row persists.
	err := p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO integ_items (id, name) VALUES ($1, $2)`,
			"tx-commit", "committed",
		)
		return err
	})
	if err != nil {
		t.Fatalf("WithTransaction (commit): %v", err)
	}
	var name string
	if err := raw.QueryRow(ctx, `SELECT name FROM integ_items WHERE id = $1`, "tx-commit").Scan(&name); err != nil {
		t.Fatalf("expected committed row to exist: %v", err)
	}
	if name != "committed" {
		t.Errorf("name: got %q, want committed", name)
	}

	// Rollback path: insert inside transaction that returns error, verify row absent.
	rollbackErr := fmt.Errorf("deliberate rollback")
	err = p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO integ_items (id, name) VALUES ($1, $2)`,
			"tx-rollback", "should not persist",
		); err != nil {
			return err
		}
		return rollbackErr
	})
	if err == nil {
		t.Fatal("expected error from failed transaction, got nil")
	}
	var count int
	if err := raw.QueryRow(ctx, `SELECT COUNT(*) FROM integ_items WHERE id = $1`, "tx-rollback").Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Errorf("expected rolled-back row to be absent, got count %d", count)
	}

	// Cleanup.
	if _, err := raw.Exec(ctx, `DELETE FROM integ_items WHERE id = $1`, "tx-commit"); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

// testRepository exercises the full CRUD cycle via the Repository interface.
func testRepository(t *testing.T) {
	ctx := context.Background()
	raw := openRawPool(t)
	defer raw.Close()

	repo := &integItemRepo{db: raw}

	t.Run("Create", func(t *testing.T) {
		item, err := repo.Create(ctx, integItem{ID: "r1", Name: "Alpha"})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if item.ID != "r1" || item.Name != "Alpha" {
			t.Errorf("got %+v, want {r1 Alpha}", item)
		}
	})

	t.Run("Get", func(t *testing.T) {
		item, err := repo.Get(ctx, "r1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if item.Name != "Alpha" {
			t.Errorf("name: got %q, want Alpha", item.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := repo.Update(ctx, integItem{ID: "r1", Name: "Beta"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Name != "Beta" {
			t.Errorf("name: got %q, want Beta", updated.Name)
		}
	})

	t.Run("List", func(t *testing.T) {
		if _, err := repo.Create(ctx, integItem{ID: "r2", Name: "Gamma"}); err != nil {
			t.Fatalf("Create second: %v", err)
		}
		items, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) < 2 {
			t.Errorf("got %d items, want at least 2", len(items))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, "r1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if err := repo.Delete(ctx, "r2"); err != nil {
			t.Fatalf("Delete r2: %v", err)
		}
		items, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List after delete: %v", err)
		}
		for _, item := range items {
			if item.ID == "r1" || item.ID == "r2" {
				t.Errorf("deleted item %q still present", item.ID)
			}
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := repo.Get(ctx, "r1")
		if err == nil {
			t.Error("expected error with cancelled context, got nil")
		}
	})
}

// testScan exercises scan.Row and scan.Rows against real query results.
func testScan(t *testing.T) {
	ctx := context.Background()
	raw := openRawPool(t)
	defer raw.Close()

	// Insert rows to scan.
	for _, item := range []integItem{{"s1", "Scan-One"}, {"s2", "Scan-Two"}} {
		if _, err := raw.Exec(ctx,
			`INSERT INTO integ_items (id, name) VALUES ($1, $2)`,
			item.ID, item.Name,
		); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	t.Cleanup(func() {
		raw.Exec(ctx, `DELETE FROM integ_items WHERE id IN ('s1','s2')`) //nolint
	})

	mapper := func(i *integItem) []any { return []any{&i.ID, &i.Name} }

	t.Run("Row", func(t *testing.T) {
		row := raw.QueryRow(ctx, `SELECT id, name FROM integ_items WHERE id = $1`, "s1")
		got, err := scan.Row(row, mapper)
		if err != nil {
			t.Fatalf("scan.Row: %v", err)
		}
		if got.ID != "s1" || got.Name != "Scan-One" {
			t.Errorf("got %+v, want {s1 Scan-One}", got)
		}
	})

	t.Run("Rows", func(t *testing.T) {
		rows, err := raw.Query(ctx,
			`SELECT id, name FROM integ_items WHERE id IN ('s1','s2') ORDER BY id`,
		)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		got, err := scan.Rows(rows, mapper)
		if err != nil {
			t.Fatalf("scan.Rows: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d rows, want 2", len(got))
		}
		if got[0].ID != "s1" || got[1].ID != "s2" {
			t.Errorf("unexpected items: %+v", got)
		}
	})
}

// testErrors verifies IsUniqueViolation and IsNotFoundError against real DB errors.
func testErrors(t *testing.T) {
	ctx := context.Background()
	raw := openRawPool(t)
	defer raw.Close()

	// Seed one row.
	if _, err := raw.Exec(ctx,
		`INSERT INTO integ_items (id, name) VALUES ($1, $2)`,
		"e1", "Error-Test",
	); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		raw.Exec(ctx, `DELETE FROM integ_items WHERE id = 'e1'`) //nolint
	})

	t.Run("UniqueViolation", func(t *testing.T) {
		_, err := raw.Exec(ctx,
			`INSERT INTO integ_items (id, name) VALUES ($1, $2)`,
			"e1", "duplicate",
		)
		if err == nil {
			t.Fatal("expected unique violation error, got nil")
		}
		if !axonerrors.IsUniqueViolation(err) {
			t.Errorf("expected IsUniqueViolation to return true, got false for: %v", err)
		}
	})

	t.Run("WrapError", func(t *testing.T) {
		inner := fmt.Errorf("inner error")
		wrapped := axonerrors.WrapError("testErrors.sub", inner)
		if wrapped.Error() != "testErrors.sub: inner error" {
			t.Errorf("unexpected wrapped message: %v", wrapped)
		}
	})

	t.Run("NotFoundError", func(t *testing.T) {
		wrapped := fmt.Errorf("get: %w", axonerrors.ErrNotFound)
		if !axonerrors.IsNotFoundError(wrapped) {
			t.Error("expected IsNotFoundError to return true for wrapped ErrNotFound")
		}
	})
}
