package pool_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/benaskins/axon-base/pool"
)

const testDSN = "postgres://aurelia:aurelia@localhost:5432/workbench?sslmode=disable"

func TestNewPool_HealthCheck(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Fatal("expected pool to be healthy")
	}
}

func TestPool_StdDB(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	db, err := p.StdDB()
	if err != nil {
		t.Fatalf("StdDB: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("StdDB ping: %v", err)
	}

	// Should return the same handle on second call.
	db2, err := p.StdDB()
	if err != nil {
		t.Fatalf("StdDB second call: %v", err)
	}
	if db != db2 {
		t.Fatal("expected StdDB to return cached handle")
	}
}

func TestPool_Close(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
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
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

	if !p.Healthy(ctx) {
		t.Skip("postgres not healthy")
	}

	m := p.Metrics()
	if m.Max == 0 {
		t.Fatal("expected Max > 0")
	}
}

func TestPool_MetricsHealthJSON(t *testing.T) {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, testDSN, "pool_test")
	if err != nil {
		t.Skip("postgres unavailable:", err)
	}
	defer p.Close()

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
