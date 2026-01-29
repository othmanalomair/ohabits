package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// GetTodosForDay retrieves todos for a specific day plus overdue incomplete todos
func (db *DB) GetTodosForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]Todo, error) {
	dateStr := date.Format("2006-01-02")

	// Get todos for this day OR overdue incomplete todos from past days
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, completed, date, created_at,
			   (date < $2 AND completed = false) as is_overdue
		FROM todos
		WHERE user_id = $1 AND is_deleted = false AND (date = $2 OR (date < $2 AND completed = false))
		ORDER BY is_overdue DESC, date ASC, created_at ASC
	`, userID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Text, &t.Completed, &t.Date, &t.CreatedAt, &t.IsOverdue); err != nil {
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
		INSERT INTO todos (user_id, text, completed, date)
		VALUES ($1, $2, false, $3)
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
		UPDATE todos SET completed = NOT completed, updated_at = NOW()
		WHERE id = $1
		RETURNING completed
	`, todoID).Scan(&completed)

	return completed, err
}

// DeleteTodo deletes a todo
func (db *DB) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `UPDATE todos SET is_deleted = true, updated_at = NOW() WHERE id = $1`, todoID)
	return err
}

// UpdateTodo updates a todo's text
func (db *DB) UpdateTodo(ctx context.Context, todoID uuid.UUID, text string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE todos SET text = $2 WHERE id = $1`, todoID, text)
	return err
}

// GetTodosForDayOnly retrieves todos only for a specific day (no overdue)
func (db *DB) GetTodosForDayOnly(ctx context.Context, userID uuid.UUID, date time.Time) ([]Todo, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, completed, date, created_at, false as is_overdue
		FROM todos
		WHERE user_id = $1 AND date = $2 AND is_deleted = false
		ORDER BY created_at ASC
	`, userID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Text, &t.Completed, &t.Date, &t.CreatedAt, &t.IsOverdue); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}

	return todos, rows.Err()
}

// UpdateTodoWithCompleted updates a todo text and completed status
func (db *DB) UpdateTodoWithCompleted(ctx context.Context, todoID uuid.UUID, text string, completed bool) error {
	_, err := db.Pool.Exec(ctx, `UPDATE todos SET text = $2, completed = $3, updated_at = NOW() WHERE id = $1`, todoID, text, completed)
	return err
}
