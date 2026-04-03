// Package scan provides helpers for mapping pgx query results to Go structs.
//
// Mapping is explicit: callers supply a RowMapper that returns field pointers
// in the same order as the SELECT columns. No reflection is used.
package scan

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

// RowMapper returns a slice of destination pointers for the given struct,
// ordered to match the query's SELECT columns.
type RowMapper[T any] func(*T) []any

// Row scans a single pgx.Row into a value of type T using mapper.
func Row[T any](row pgx.Row, mapper RowMapper[T]) (T, error) {
	var t T
	if err := row.Scan(mapper(&t)...); err != nil {
		return t, fmt.Errorf("scan.Row: %w", err)
	}
	return t, nil
}

// Rows scans all rows into a slice of T using mapper.
// The rows are closed before returning.
func Rows[T any](rows pgx.Rows, mapper RowMapper[T]) ([]T, error) {
	defer rows.Close()
	var result []T
	for rows.Next() {
		var t T
		if err := rows.Scan(mapper(&t)...); err != nil {
			return nil, fmt.Errorf("scan.Rows: %w", err)
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan.Rows: %w", err)
	}
	return result, nil
}
