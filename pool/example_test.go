package pool_test

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/benaskins/axon-base/pool"
)

func ExampleNewPool() {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, "postgres://postgres@localhost:5432/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	if p.Healthy(ctx) {
		log.Println("database is reachable")
	}
}

func ExamplePool_WithTransaction() {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, "postgres://postgres@localhost:5432/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	err = p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO users (id, name) VALUES ($1, $2)", "abc", "alice")
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExamplePool_Metrics() {
	ctx := context.Background()
	p, err := pool.NewPool(ctx, "postgres://postgres@localhost:5432/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	m := p.Metrics()
	data, err := m.HealthJSON()
	if err != nil {
		log.Fatal(err)
	}
	// data is JSON suitable for a health endpoint:
	// {"active":0,"idle":0,"total":0,"max":4,"wait_time_ms":0}
	_ = data
}
