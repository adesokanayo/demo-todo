package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Postgres stores todos in a real Postgres database.
type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (s *Postgres) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS todos (
			id         BIGSERIAL PRIMARY KEY,
			title      TEXT NOT NULL CHECK (length(btrim(title)) > 0),
			done       BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("migrate todos: %w", err)
	}
	return nil
}

func (s *Postgres) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Postgres) List(ctx context.Context) ([]Todo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, done, created_at
		FROM todos
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var item Todo
		if err := rows.Scan(&item.ID, &item.Title, &item.Done, &item.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, item)
	}
	return todos, rows.Err()
}

func (s *Postgres) Create(ctx context.Context, title string) (Todo, error) {
	var item Todo
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, done, created_at`, strings.TrimSpace(title)).
		Scan(&item.ID, &item.Title, &item.Done, &item.CreatedAt)
	if err != nil {
		return Todo{}, err
	}
	return item, nil
}

func (s *Postgres) Get(ctx context.Context, id int64) (Todo, error) {
	var item Todo
	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, done, created_at
		FROM todos
		WHERE id = $1`, id).
		Scan(&item.ID, &item.Title, &item.Done, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, ErrNotFound
	}
	if err != nil {
		return Todo{}, err
	}
	return item, nil
}

func (s *Postgres) Update(ctx context.Context, id int64, in UpdateTodoRequest) (Todo, error) {
	var title *string
	if in.Title != nil {
		trimmed := strings.TrimSpace(*in.Title)
		title = &trimmed
	}
	var done *bool
	if in.Done != nil {
		value := *in.Done
		done = &value
	}

	var item Todo
	err := s.db.QueryRowContext(ctx, `
		UPDATE todos
		SET title = COALESCE($2, title),
		    done = COALESCE($3, done)
		WHERE id = $1
		RETURNING id, title, done, created_at`, id, title, done).
		Scan(&item.ID, &item.Title, &item.Done, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, ErrNotFound
	}
	if err != nil {
		return Todo{}, err
	}
	return item, nil
}
