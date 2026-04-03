package pool_test

import (
	"context"
	"testing"

	"github.com/benaskins/axon-base/pool"
)

const testDSN = "postgres://postgres@localhost:5432/workbench"

func TestNewPool_HealthCheck(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Skip("postgres unavailable")
	}
}

func TestPool_Close(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}

	p.Close()

	if p.Healthy(ctx) {
		t.Fatal("expected Healthy to return false after Close")
	}
}

func TestPool_Metrics(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	// Trigger a connection by pinging.
	if !p.Healthy(ctx) {
		t.Skip("postgres not healthy")
	}

	m := p.Metrics()
	if m.MaxConns == 0 {
		t.Fatal("expected MaxConns > 0")
	}
}
