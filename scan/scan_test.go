package scan_test

import (
	"context"
	"testing"

	"github.com/benaskins/axon-base/scan"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testDSN = "postgres://postgres@localhost:5432/workbench"

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	db, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	if err := db.Ping(ctx); err != nil {
		t.Skip("postgres unavailable:", err)
	}
	t.Cleanup(db.Close)
	return db
}

type item struct {
	ID   int
	Name string
}

func itemMapper(i *item) []any {
	return []any{&i.ID, &i.Name}
}

func TestScanRow(t *testing.T) {
	db := openPool(t)
	ctx := context.Background()

	row := db.QueryRow(ctx, "SELECT 42, 'hello'")
	got, err := scan.Row(row, itemMapper)
	if err != nil {
		t.Fatalf("scan.Row: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("ID: got %d, want 42", got.ID)
	}
	if got.Name != "hello" {
		t.Errorf("Name: got %q, want %q", got.Name, "hello")
	}
}

func TestScanRows(t *testing.T) {
	db := openPool(t)
	ctx := context.Background()

	rows, err := db.Query(ctx, "SELECT n, 'item-' || n FROM generate_series(1, 3) AS n")
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	got, err := scan.Rows(rows, itemMapper)
	if err != nil {
		t.Fatalf("scan.Rows: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len: got %d, want 3", len(got))
	}
	for i, item := range got {
		wantID := i + 1
		wantName := "item-" + string(rune('0'+wantID))
		if item.ID != wantID {
			t.Errorf("[%d] ID: got %d, want %d", i, item.ID, wantID)
		}
		if item.Name != wantName {
			t.Errorf("[%d] Name: got %q, want %q", i, item.Name, wantName)
		}
	}
}

func TestScanRows_Empty(t *testing.T) {
	db := openPool(t)
	ctx := context.Background()

	rows, err := db.Query(ctx, "SELECT n, 'x' FROM generate_series(1, 0) AS n")
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	got, err := scan.Rows(rows, itemMapper)
	if err != nil {
		t.Fatalf("scan.Rows: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil slice for empty result, got %v", got)
	}
}
