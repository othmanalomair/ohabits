package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// GetTodosForDay retrieves todos for a specific day
func (db *DB) GetTodosForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]Todo, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, completed, date, created_at
		FROM todos
		WHERE user_id = $1 AND date = $2
		ORDER BY created_at
	`, userID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Text, &t.Completed, &t.Date, &t.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}

	return todos, rows.Err()
}

// CreateTodo creates a new todo
func (db *DB) CreateTodo(ctx context.Context, userID uuid.UUID, text string, date time.Time) (*Todo, error) {
	dateStr := date.Format("2006-01-02")

	var t Todo
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO todos (user_id, text, date)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, text, completed, date, created_at
	`, userID, text, dateStr).Scan(&t.ID, &t.UserID, &t.Text, &t.Completed, &t.Date, &t.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

// ToggleTodo toggles the completion status of a todo
func (db *DB) ToggleTodo(ctx context.Context, todoID uuid.UUID) (bool, error) {
	var completed bool
	err := db.Pool.QueryRow(ctx, `
		UPDATE todos SET completed = NOT completed
		WHERE id = $1
		RETURNING completed
	`, todoID).Scan(&completed)

	return completed, err
}

// DeleteTodo deletes a todo
func (db *DB) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, todoID)
	return err
}

// UpdateTodo updates a todo's text
func (db *DB) UpdateTodo(ctx context.Context, todoID uuid.UUID, text string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE todos SET text = $2 WHERE id = $1`, todoID, text)
	return err
}
