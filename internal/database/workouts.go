package database

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetWorkoutsByUserID retrieves all workouts for a user
func (db *DB) GetWorkoutsByUserID(ctx context.Context, userID uuid.UUID) ([]Workout, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, day, exercises, display_order, is_active, created_at, updated_at
		FROM workouts
		WHERE user_id = $1
		ORDER BY display_order
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		var exercisesJSON []byte
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesJSON,
			&w.DisplayOrder, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(exercisesJSON, &w.Exercises)
		workouts = append(workouts, w)
	}

	return workouts, rows.Err()
}

// GetWorkoutsForDay retrieves workouts scheduled for a specific day
func (db *DB) GetWorkoutsForDay(ctx context.Context, userID uuid.UUID, day string) ([]Workout, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, day, exercises, display_order, is_active, created_at, updated_at
		FROM workouts
		WHERE user_id = $1 AND (day = $2 OR day = '' OR day IS NULL) AND is_active = true
		ORDER BY display_order
	`, userID, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		var exercisesJSON []byte
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesJSON,
			&w.DisplayOrder, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(exercisesJSON, &w.Exercises)
		workouts = append(workouts, w)
	}

	return workouts, rows.Err()
}

// CreateWorkout creates a new workout
func (db *DB) CreateWorkout(ctx context.Context, userID uuid.UUID, name, day string, exercises []Exercise) (*Workout, error) {
	exercisesJSON, _ := json.Marshal(exercises)

	var w Workout
	var exercisesBytes []byte
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO workouts (user_id, name, day, exercises)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, day, exercises, display_order, is_active, created_at, updated_at
	`, userID, name, day, exercisesJSON).Scan(
		&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesBytes,
		&w.DisplayOrder, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(exercisesBytes, &w.Exercises)
	return &w, nil
}

// GetWorkoutLogForDay retrieves workout log for a specific day
func (db *DB) GetWorkoutLogForDay(ctx context.Context, userID uuid.UUID, date time.Time) (*WorkoutLog, error) {
	dateStr := date.Format("2006-01-02")

	var wl WorkoutLog
	var exercisesJSON, cardioJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, workout_name, completed_exercises, cardio, weight, date, notes, created_at
		FROM workout_logs
		WHERE user_id = $1 AND date = $2
	`, userID, dateStr).Scan(
		&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesJSON, &cardioJSON,
		&wl.Weight, &wl.Date, &wl.Notes, &wl.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal(exercisesJSON, &wl.CompletedExercises)
	if cardioJSON != nil {
		json.Unmarshal(cardioJSON, &wl.Cardio)
	}

	return &wl, nil
}

// SaveWorkoutLog creates or updates a workout log
func (db *DB) SaveWorkoutLog(ctx context.Context, userID uuid.UUID, workoutName string, exercises []Exercise, cardio *Cardio, weight float64, date time.Time, notes string) (*WorkoutLog, error) {
	dateStr := date.Format("2006-01-02")
	exercisesJSON, _ := json.Marshal(exercises)

	var cardioJSON []byte
	if cardio != nil {
		cardioJSON, _ = json.Marshal(cardio)
	}

	// Check if log exists
	var existingID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		SELECT id FROM workout_logs WHERE user_id = $1 AND date = $2
	`, userID, dateStr).Scan(&existingID)

	var wl WorkoutLog
	var exercisesBytes, cardioBytes []byte

	if err == nil {
		// Update existing
		err = db.Pool.QueryRow(ctx, `
			UPDATE workout_logs
			SET workout_name = $2, completed_exercises = $3, cardio = $4, weight = $5, notes = $6
			WHERE id = $1
			RETURNING id, user_id, workout_name, completed_exercises, cardio, weight, date, notes, created_at
		`, existingID, workoutName, exercisesJSON, cardioJSON, weight, notes).Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesBytes, &cardioBytes,
			&wl.Weight, &wl.Date, &wl.Notes, &wl.CreatedAt,
		)
	} else {
		// Create new
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO workout_logs (user_id, workout_name, completed_exercises, cardio, weight, date, notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, user_id, workout_name, completed_exercises, cardio, weight, date, notes, created_at
		`, userID, workoutName, exercisesJSON, cardioJSON, weight, dateStr, notes).Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesBytes, &cardioBytes,
			&wl.Weight, &wl.Date, &wl.Notes, &wl.CreatedAt,
		)
	}

	if err != nil {
		return nil, err
	}

	json.Unmarshal(exercisesBytes, &wl.CompletedExercises)
	if cardioBytes != nil {
		json.Unmarshal(cardioBytes, &wl.Cardio)
	}

	return &wl, nil
}
