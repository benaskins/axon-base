package pool_test

import (
	"context"
	"encoding/json"
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
	if m.Max == 0 {
		t.Fatal("expected Max > 0")
	}
	// WaitTime must be a valid duration (zero is fine for an idle pool).
	_ = m.WaitTime
}

func TestPool_MetricsHealthJSON(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN)
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Skip("postgres not healthy")
	}

	data, err := p.Metrics().HealthJSON()
	if err != nil {
		t.Fatalf("HealthJSON: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, key := range []string{"active", "idle", "total", "max", "wait_time_ms"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing key %q in health JSON", key)
		}
	}
}
