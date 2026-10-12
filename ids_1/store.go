package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *Store) List(ctx context.Context) ([]Task, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, title, done, created_at FROM tasks ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) Get(ctx context.Context, id int64) (Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (s *Store) Create(ctx context.Context, title string) (Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tasks (title) VALUES ($1) RETURNING id, title, done, created_at`, title).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	return t, err
}

func (s *Store) SetDone(ctx context.Context, id int64, done bool) (Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		`UPDATE tasks SET done = $2 WHERE id = $1 RETURNING id, title, done, created_at`, id, done).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
