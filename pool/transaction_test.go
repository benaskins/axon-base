package pool_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/benaskins/axon-base/pool"
)

func TestWithTransaction_Commit(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Skip("postgres unavailable")
	}

	var called bool
	err = p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !called {
		t.Fatal("expected callback to be called")
	}
}

func TestWithTransaction_Rollback(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Skip("postgres unavailable")
	}

	callbackErr := errors.New("something went wrong")
	err = p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return callbackErr
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected error to wrap callback error, got: %v", err)
	}
}

func TestWithTransaction_ErrorWrapped(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Skip("postgres unavailable")
	}

	sentinel := errors.New("sentinel")
	err = p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return fmt.Errorf("inner: %w", sentinel)
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected error chain to contain sentinel, got: %v", err)
	}
}
