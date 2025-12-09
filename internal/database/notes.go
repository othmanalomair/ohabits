package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetNoteForDay retrieves note for a specific day
func (db *DB) GetNoteForDay(ctx context.Context, userID uuid.UUID, date time.Time) (*Note, error) {
	dateStr := date.Format("2006-01-02")

	var n Note
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, text, date, created_at, updated_at
		FROM notes
		WHERE user_id = $1 AND date = $2
	`, userID, dateStr).Scan(&n.ID, &n.UserID, &n.Text, &n.Date, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No note for this day
		}
		return nil, err
	}

	return &n, nil
}

// SaveNote creates or updates a note for a specific day
func (db *DB) SaveNote(ctx context.Context, userID uuid.UUID, text string, date time.Time) (*Note, error) {
	dateStr := date.Format("2006-01-02")

	var n Note
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO notes (user_id, text, date)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, date) DO UPDATE SET text = $2, updated_at = NOW()
		RETURNING id, user_id, text, date, created_at, updated_at
	`, userID, text, dateStr).Scan(&n.ID, &n.UserID, &n.Text, &n.Date, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &n, nil
}

// GetMoodForDay retrieves mood rating for a specific day
func (db *DB) GetMoodForDay(ctx context.Context, userID uuid.UUID, date time.Time) (*MoodRating, error) {
	dateStr := date.Format("2006-01-02")

	var m MoodRating
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, rating, date, created_at
		FROM mood_ratings
		WHERE user_id = $1 AND date = $2
	`, userID, dateStr).Scan(&m.ID, &m.UserID, &m.Rating, &m.Date, &m.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}

// SaveMood creates or updates mood rating for a specific day
func (db *DB) SaveMood(ctx context.Context, userID uuid.UUID, rating int, date time.Time) (*MoodRating, error) {
	dateStr := date.Format("2006-01-02")

	var m MoodRating
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO mood_ratings (user_id, rating, date)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, date) DO UPDATE SET rating = $2
		RETURNING id, user_id, rating, date, created_at
	`, userID, rating, dateStr).Scan(&m.ID, &m.UserID, &m.Rating, &m.Date, &m.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &m, nil
}
