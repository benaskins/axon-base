package pool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TxFunc is a callback executed within a transaction.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// WithTransaction begins a transaction, executes fn, and commits on success.
// If fn returns an error, the transaction is rolled back and the error is returned.
func (p *Pool) WithTransaction(ctx context.Context, fn TxFunc) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.WithTransaction: begin: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("pool.WithTransaction: commit: %w", err)
	}

	return nil
}
