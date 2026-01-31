package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GetHabitsByUserID retrieves all habits for a user
func (db *DB) GetHabitsByUserID(ctx context.Context, userID uuid.UUID) ([]Habit, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, icon, scheduled_days, created_at, updated_at, COALESCE(is_deleted, false) as is_deleted
		FROM habits WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []Habit
	for rows.Next() {
		var h Habit
		var daysJSON []byte
		if err := rows.Scan(&h.ID, &h.UserID, &h.Name, &h.Icon, &daysJSON, &h.CreatedAt, &h.UpdatedAt, &h.IsDeleted); err != nil {
			return nil, err
		}
		json.Unmarshal(daysJSON, &h.ScheduledDays)
		habits = append(habits, h)
	}

	return habits, rows.Err()
}

// GetHabitsForDay retrieves habits scheduled for a specific day with completion status
func (db *DB) GetHabitsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]HabitWithCompletion, error) {
	weekday := date.Weekday()

	rows, err := db.Pool.Query(ctx, `
		SELECT h.id, h.user_id, h.name, h.icon, h.scheduled_days, h.created_at, h.updated_at,
			   COALESCE(hc.completed, false) as completed
		FROM habits h
		LEFT JOIN habits_completions hc ON h.id = hc.habit_id AND hc.date = $2
		WHERE h.user_id = $1 AND COALESCE(h.is_deleted, false) = false
		ORDER BY h.created_at
	`, userID, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []HabitWithCompletion
	for rows.Next() {
		var h HabitWithCompletion
		var daysJSON []byte
		if err := rows.Scan(&h.ID, &h.UserID, &h.Name, &h.Icon, &daysJSON, &h.CreatedAt, &h.UpdatedAt, &h.Completed); err != nil {
			return nil, err
		}
		json.Unmarshal(daysJSON, &h.ScheduledDays)

		// Check if habit is scheduled for this day
		if h.IsScheduledFor(weekday) {
			habits = append(habits, h)
		}
	}

	return habits, rows.Err()
}

// CreateHabit creates a new habit
func (db *DB) CreateHabit(ctx context.Context, userID uuid.UUID, name string, icon string, scheduledDays []string) (*Habit, error) {
	daysJSON, _ := json.Marshal(scheduledDays)
	if icon == "" {
		icon = "checkmark.circle.fill"
	}

	var h Habit
	var daysBytes []byte
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO habits (user_id, name, icon, scheduled_days)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, icon, scheduled_days, created_at, updated_at
	`, userID, name, icon, daysJSON).Scan(&h.ID, &h.UserID, &h.Name, &h.Icon, &daysBytes, &h.CreatedAt, &h.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(daysBytes, &h.ScheduledDays)
	return &h, nil
}

// UpdateHabit updates a habit
func (db *DB) UpdateHabit(ctx context.Context, habitID uuid.UUID, name string, icon string, scheduledDays []string) error {
	daysJSON, _ := json.Marshal(scheduledDays)
	if icon == "" {
		icon = "checkmark.circle.fill"
	}

	_, err := db.Pool.Exec(ctx, `
		UPDATE habits SET name = $2, icon = $3, scheduled_days = $4, updated_at = NOW()
		WHERE id = $1
	`, habitID, name, icon, daysJSON)

	return err
}

// DeleteHabit deletes a habit
func (db *DB) DeleteHabit(ctx context.Context, habitID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM habits WHERE id = $1`, habitID)
	return err
}

// ToggleHabitCompletion toggles the completion status of a habit for a date
func (db *DB) ToggleHabitCompletion(ctx context.Context, userID, habitID uuid.UUID, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")

	// Check current status
	var completed bool
	err := db.Pool.QueryRow(ctx, `
		SELECT completed FROM habits_completions
		WHERE habit_id = $1 AND date = $2
	`, habitID, dateStr).Scan(&completed)

	if err != nil {
		// No record exists, create one with completed = true
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO habits_completions (habit_id, user_id, completed, date)
			VALUES ($1, $2, true, $3)
		`, habitID, userID, dateStr)
		return true, err
	}

	// Toggle existing record
	newStatus := !completed
	_, err = db.Pool.Exec(ctx, `
		UPDATE habits_completions SET completed = $3, updated_at = NOW()
		WHERE habit_id = $1 AND date = $2
	`, habitID, dateStr, newStatus)

	return newStatus, err
}

// UpsertHabitCompletion sets the habit completion status directly (for sync)
func (db *DB) UpsertHabitCompletion(ctx context.Context, userID, habitID uuid.UUID, date time.Time, completed bool) error {
	dateStr := date.Format("2006-01-02")

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO habits_completions (habit_id, user_id, completed, date)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (habit_id, date)
		DO UPDATE SET completed = $3
	`, habitID, userID, completed, dateStr)

	return err
}

// SoftDeleteHabit marks a habit as deleted (for sync)
func (db *DB) SoftDeleteHabit(ctx context.Context, habitID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE habits SET is_deleted = true, updated_at = NOW()
		WHERE id = $1
	`, habitID)
	return err
}
