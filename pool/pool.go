// Package pool provides a pgxpool.Pool wrapper with schema isolation,
// database/sql compatibility, health check, and graceful shutdown.
package pool

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Pool wraps pgxpool.Pool with schema isolation and database/sql compatibility.
type Pool struct {
	db     *pgxpool.Pool
	stdDB  *sql.DB
	schema string
	dsn    string
}

// NewPool opens a connection pool for the given DSN with schema isolation.
// It creates the schema if it doesn't exist and sets the search_path.
func NewPool(ctx context.Context, dsn, schema string) (*Pool, error) {
	// Create schema via a temporary connection.
	tmpDB, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pool: open: %w", err)
	}
	createSQL := "CREATE SCHEMA IF NOT EXISTS " + pgx.Identifier{schema}.Sanitize()
	if _, err := tmpDB.Exec(ctx, createSQL); err != nil {
		tmpDB.Close()
		return nil, fmt.Errorf("pool: create schema %s: %w", schema, err)
	}
	tmpDB.Close()

	// Reopen with search_path baked in.
	dsnWithSchema := appendSearchPath(dsn, schema)
	db, err := pgxpool.New(ctx, dsnWithSchema)
	if err != nil {
		return nil, fmt.Errorf("pool: open with search_path: %w", err)
	}

	return &Pool{db: db, schema: schema, dsn: dsnWithSchema}, nil
}

// StdDB returns a *sql.DB handle backed by the same DSN and search_path.
// The handle is created lazily and cached for the lifetime of the Pool.
// This provides compatibility with libraries that require database/sql
// (e.g., goose migrations, axon-fact event stores).
func (p *Pool) StdDB() (*sql.DB, error) {
	if p.stdDB != nil {
		return p.stdDB, nil
	}
	db, err := sql.Open("pgx", p.dsn)
	if err != nil {
		return nil, fmt.Errorf("pool: open std db: %w", err)
	}
	p.stdDB = db
	return db, nil
}

// Healthy pings the database and returns true if it responds.
func (p *Pool) Healthy(ctx context.Context) bool {
	return p.db.Ping(ctx) == nil
}

// Close closes all connections in the pool and the database/sql handle if opened.
func (p *Pool) Close() {
	p.db.Close()
	if p.stdDB != nil {
		p.stdDB.Close()
	}
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

// appendSearchPath adds search_path to a PostgreSQL DSN via the options parameter.
func appendSearchPath(dsn, schema string) string {
	opt := fmt.Sprintf("-csearch_path=%s", schema)
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return dsn
		}
		q := u.Query()
		q.Set("options", opt)
		u.RawQuery = q.Encode()
		return u.String()
	}
	return dsn + " options=" + opt
}
