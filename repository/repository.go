// Package repository defines the Repository interface pattern for CRUD operations.
package repository

import "context"

// Repository is a generic contract for data access.
type Repository[T any] interface {
	Create(ctx context.Context, entity T) (T, error)
	Get(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, entity T) (T, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]T, error)
}
