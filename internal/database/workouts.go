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
		SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at
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
			&w.DisplayOrder, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(exercisesJSON, &w.Exercises)
		workouts = append(workouts, w)
	}

	return workouts, rows.Err()
}

// GetWorkoutsForDay retrieves workouts scheduled for a specific day
// Note: Currently returns all workouts regardless of day since day filtering is not needed
func (db *DB) GetWorkoutsForDay(ctx context.Context, userID uuid.UUID, day string) ([]Workout, error) {
	return db.GetWorkoutsByUserID(ctx, userID)
}

// CreateWorkout creates a new workout
func (db *DB) CreateWorkout(ctx context.Context, userID uuid.UUID, name, day string, exercises []Exercise) (*Workout, error) {
	exercisesJSON, _ := json.Marshal(exercises)

	// Get next display_order for this user
	var nextOrder int
	err := db.Pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(display_order), 0) + 1 FROM workouts WHERE user_id = $1
	`, userID).Scan(&nextOrder)
	if err != nil {
		return nil, err
	}

	var w Workout
	var exercisesBytes []byte
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO workouts (user_id, name, day, exercises, display_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, day, exercises, display_order, created_at, updated_at
	`, userID, name, day, exercisesJSON, nextOrder).Scan(
		&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesBytes,
		&w.DisplayOrder, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(exercisesBytes, &w.Exercises)
	return &w, nil
}

// GetWorkoutByID retrieves a single workout by ID
func (db *DB) GetWorkoutByID(ctx context.Context, workoutID uuid.UUID) (*Workout, error) {
	var w Workout
	var exercisesJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at
		FROM workouts WHERE id = $1
	`, workoutID).Scan(
		&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesJSON,
		&w.DisplayOrder, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(exercisesJSON, &w.Exercises)
	return &w, nil
}

// UpdateWorkout updates a workout
func (db *DB) UpdateWorkout(ctx context.Context, workoutID uuid.UUID, name, day string, exercises []Exercise) error {
	exercisesJSON, _ := json.Marshal(exercises)

	_, err := db.Pool.Exec(ctx, `
		UPDATE workouts
		SET name = $2, day = $3, exercises = $4, updated_at = now()
		WHERE id = $1
	`, workoutID, name, day, exercisesJSON)

	return err
}

// DeleteWorkout deletes a workout
func (db *DB) DeleteWorkout(ctx context.Context, workoutID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM workouts WHERE id = $1`, workoutID)
	return err
}

// ReorderWorkouts updates the display_order of workouts based on the provided order
func (db *DB) ReorderWorkouts(ctx context.Context, workoutIDs []uuid.UUID) error {
	for i, id := range workoutIDs {
		_, err := db.Pool.Exec(ctx, `
			UPDATE workouts SET display_order = $1, updated_at = now() WHERE id = $2
		`, i+1, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetWorkoutLogForDay retrieves workout log for a specific day
func (db *DB) GetWorkoutLogForDay(ctx context.Context, userID uuid.UUID, date time.Time) (*WorkoutLog, error) {
	dateStr := date.Format("2006-01-02")

	var wl WorkoutLog
	var exercisesJSON, cardioJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at
		FROM workout_logs
		WHERE user_id = $1 AND date = $2
	`, userID, dateStr).Scan(
		&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesJSON, &cardioJSON,
		&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt,
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
func (db *DB) SaveWorkoutLog(ctx context.Context, userID uuid.UUID, workoutName string, exercises []Exercise, cardio []Cardio, weight float64, date time.Time) (*WorkoutLog, error) {
	dateStr := date.Format("2006-01-02")
	exercisesJSON, _ := json.Marshal(exercises)
	cardioJSON, _ := json.Marshal(cardio)

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
			SET name = $2, completed_exercises = $3, cardio = $4, weight = $5, updated_at = now()
			WHERE id = $1
			RETURNING id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at
		`, existingID, workoutName, exercisesJSON, cardioJSON, weight).Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesBytes, &cardioBytes,
			&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt,
		)
	} else {
		// Create new
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO workout_logs (user_id, name, completed_exercises, cardio, weight, date)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at
		`, userID, workoutName, exercisesJSON, cardioJSON, weight, dateStr).Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesBytes, &cardioBytes,
			&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt,
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
