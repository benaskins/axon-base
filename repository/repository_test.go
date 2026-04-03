package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/benaskins/axon-base/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testItem is a simple entity used to exercise the Repository interface.
type testItem struct {
	ID   string
	Name string
}

// testItemRepo is a concrete Repository[testItem] backed by Postgres.
// It demonstrates the intended implementation pattern for axon-base users:
// embed a pool, write explicit SQL, return wrapped errors.
type testItemRepo struct {
	pool *pgxpool.Pool
}

// Compile-time check: testItemRepo satisfies Repository[testItem].
var _ repository.Repository[testItem] = (*testItemRepo)(nil)

func (r *testItemRepo) Create(ctx context.Context, item testItem) (testItem, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO test_items (id, name) VALUES ($1, $2) RETURNING id, name`,
		item.ID, item.Name,
	)
	var out testItem
	if err := row.Scan(&out.ID, &out.Name); err != nil {
		return testItem{}, fmt.Errorf("testItemRepo.Create: %w", err)
	}
	return out, nil
}

func (r *testItemRepo) Get(ctx context.Context, id string) (testItem, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name FROM test_items WHERE id = $1`,
		id,
	)
	var item testItem
	if err := row.Scan(&item.ID, &item.Name); err != nil {
		return testItem{}, fmt.Errorf("testItemRepo.Get: %w", err)
	}
	return item, nil
}

func (r *testItemRepo) Update(ctx context.Context, item testItem) (testItem, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE test_items SET name = $2 WHERE id = $1 RETURNING id, name`,
		item.ID, item.Name,
	)
	var out testItem
	if err := row.Scan(&out.ID, &out.Name); err != nil {
		return testItem{}, fmt.Errorf("testItemRepo.Update: %w", err)
	}
	return out, nil
}

func (r *testItemRepo) Delete(ctx context.Context, id string) error {
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM test_items WHERE id = $1`, id,
	); err != nil {
		return fmt.Errorf("testItemRepo.Delete: %w", err)
	}
	return nil
}

func (r *testItemRepo) List(ctx context.Context) ([]testItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM test_items ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("testItemRepo.List: %w", err)
	}
	defer rows.Close()
	var items []testItem
	for rows.Next() {
		var item testItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, fmt.Errorf("testItemRepo.List: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("testItemRepo.List: %w", err)
	}
	return items, nil
}

// openTestPool connects to the local workbench DB and skips if unavailable.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgresql://postgres@localhost:5432/workbench")
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("postgres unavailable: %v", err)
	}
	return pool
}

func TestRepositoryCRUD(t *testing.T) {
	pool := openTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_items (
			id   TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE test_items`); err != nil {
		t.Fatalf("truncate table: %v", err)
	}
	t.Cleanup(func() {
		conn, err := pgxpool.New(context.Background(), "postgresql://postgres@localhost:5432/workbench")
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_items`)
	})

	repo := &testItemRepo{pool: pool}

	t.Run("Create", func(t *testing.T) {
		item, err := repo.Create(ctx, testItem{ID: "1", Name: "Alice"})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if item.ID != "1" || item.Name != "Alice" {
			t.Errorf("got %+v, want {1 Alice}", item)
		}
	})

	t.Run("Get", func(t *testing.T) {
		item, err := repo.Get(ctx, "1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if item.Name != "Alice" {
			t.Errorf("got name %q, want Alice", item.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := repo.Update(ctx, testItem{ID: "1", Name: "Bob"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Name != "Bob" {
			t.Errorf("got name %q, want Bob", updated.Name)
		}
	})

	t.Run("List", func(t *testing.T) {
		if _, err := repo.Create(ctx, testItem{ID: "2", Name: "Carol"}); err != nil {
			t.Fatalf("Create second item: %v", err)
		}
		items, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("got %d items, want 2", len(items))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, "1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		items, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List after delete: %v", err)
		}
		if len(items) != 1 {
			t.Errorf("got %d items after delete, want 1", len(items))
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := repo.Get(ctx, "2")
		if err == nil {
			t.Error("expected error with cancelled context, got nil")
		}
	})
}
