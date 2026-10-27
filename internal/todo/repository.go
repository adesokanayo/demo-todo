package todo

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a todo id does not exist.
var ErrNotFound = errors.New("todo not found")

// Repository is the persistence boundary used by the HTTP handlers.
type Repository interface {
	Ping(ctx context.Context) error
	List(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Get(ctx context.Context, id int64) (Todo, error)
	Update(ctx context.Context, id int64, in UpdateTodoRequest) (Todo, error)
}
