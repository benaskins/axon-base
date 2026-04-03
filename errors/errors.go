// Package errors provides error wrapping utilities for pgx database operations.
package errors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// WrapError wraps err with the given operation context. Returns nil if err is nil.
func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// IsNotFoundError reports whether err is or wraps ErrNotFound.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUniqueViolation reports whether err is or wraps a pgx unique constraint violation (SQLSTATE 23505).
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
