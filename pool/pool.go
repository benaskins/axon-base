// Package pool provides a pgxpool.Pool wrapper with health check and graceful shutdown.
package pool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool.
type Pool struct {
	db *pgxpool.Pool
}

// NewPool opens a connection pool for the given DSN.
func NewPool(ctx context.Context, dsn string) (*Pool, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pool.NewPool: %w", err)
	}
	return &Pool{db: db}, nil
}

// Healthy pings the database and returns true if it responds.
func (p *Pool) Healthy(ctx context.Context) bool {
	return p.db.Ping(ctx) == nil
}

// Close closes all connections in the pool.
func (p *Pool) Close() {
	p.db.Close()
}

// Metrics holds a snapshot of pool statistics.
type Metrics struct {
	AcquiredConns int32
	IdleConns     int32
	TotalConns    int32
	MaxConns      int32
}

// Metrics returns a snapshot of the current pool statistics.
func (p *Pool) Metrics() Metrics {
	s := p.db.Stat()
	return Metrics{
		AcquiredConns: s.AcquiredConns(),
		IdleConns:     s.IdleConns(),
		TotalConns:    s.TotalConns(),
		MaxConns:      s.MaxConns(),
	}
}
