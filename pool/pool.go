// Package pool provides a pgxpool.Pool wrapper with health check and graceful shutdown.
package pool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	Active   int32         `json:"active"`
	Idle     int32         `json:"idle"`
	Total    int32         `json:"total"`
	Max      int32         `json:"max"`
	WaitTime time.Duration `json:"-"`
}

// HealthJSON returns a JSON-encoded representation of the metrics suitable
// for use in health endpoints. WaitTime is expressed as milliseconds.
func (m Metrics) HealthJSON() ([]byte, error) {
	return json.Marshal(struct {
		Active     int32   `json:"active"`
		Idle       int32   `json:"idle"`
		Total      int32   `json:"total"`
		Max        int32   `json:"max"`
		WaitTimeMS float64 `json:"wait_time_ms"`
	}{
		Active:     m.Active,
		Idle:       m.Idle,
		Total:      m.Total,
		Max:        m.Max,
		WaitTimeMS: float64(m.WaitTime) / float64(time.Millisecond),
	})
}

// Metrics returns a snapshot of the current pool statistics.
func (p *Pool) Metrics() Metrics {
	s := p.db.Stat()
	return Metrics{
		Active:   s.AcquiredConns(),
		Idle:     s.IdleConns(),
		Total:    s.TotalConns(),
		Max:      s.MaxConns(),
		WaitTime: s.EmptyAcquireWaitTime(),
	}
}
