package repository_test

import (
	"context"

	"github.com/benaskins/axon-base/repository"
)

// Widget is an example domain type.
type Widget struct {
	ID    string
	Label string
}

// widgetRepo is a concrete implementation of Repository[Widget].
// In practice it would hold a *pool.Pool and execute explicit SQL.
type widgetRepo struct{}

func (r *widgetRepo) Create(ctx context.Context, w Widget) (Widget, error) {
	// INSERT INTO widgets (id, label) VALUES ($1, $2) RETURNING id, label
	return w, nil
}

func (r *widgetRepo) Get(ctx context.Context, id string) (Widget, error) {
	// SELECT id, label FROM widgets WHERE id = $1
	return Widget{ID: id}, nil
}

func (r *widgetRepo) Update(ctx context.Context, w Widget) (Widget, error) {
	// UPDATE widgets SET label = $1 WHERE id = $2 RETURNING id, label
	return w, nil
}

func (r *widgetRepo) Delete(ctx context.Context, id string) error {
	// DELETE FROM widgets WHERE id = $1
	return nil
}

func (r *widgetRepo) List(ctx context.Context) ([]Widget, error) {
	// SELECT id, label FROM widgets ORDER BY label
	return nil, nil
}

// Compile-time assertion: widgetRepo satisfies Repository[Widget].
var _ repository.Repository[Widget] = (*widgetRepo)(nil)
