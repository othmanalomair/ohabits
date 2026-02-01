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
		SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at, is_rest_day
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
			&w.DisplayOrder, &w.CreatedAt, &w.UpdatedAt, &w.IsRestDay,
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
	return db.CreateWorkoutWithRestDay(ctx, userID, name, day, exercises, false)
}

// CreateWorkoutWithRestDay creates a new workout with rest day flag
func (db *DB) CreateWorkoutWithRestDay(ctx context.Context, userID uuid.UUID, name, day string, exercises []Exercise, isRestDay bool) (*Workout, error) {
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
		INSERT INTO workouts (user_id, name, day, exercises, display_order, is_rest_day)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, name, day, exercises, display_order, created_at, updated_at, is_rest_day
	`, userID, name, day, exercisesJSON, nextOrder, isRestDay).Scan(
		&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesBytes,
		&w.DisplayOrder, &w.CreatedAt, &w.UpdatedAt, &w.IsRestDay,
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
		SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at, is_rest_day
		FROM workouts WHERE id = $1
	`, workoutID).Scan(
		&w.ID, &w.UserID, &w.Name, &w.Day, &exercisesJSON,
		&w.DisplayOrder, &w.CreatedAt, &w.UpdatedAt, &w.IsRestDay,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(exercisesJSON, &w.Exercises)
	return &w, nil
}

// UpdateWorkout updates a workout (without changing is_rest_day)
func (db *DB) UpdateWorkout(ctx context.Context, workoutID uuid.UUID, name, day string, exercises []Exercise) error {
	exercisesJSON, _ := json.Marshal(exercises)

	_, err := db.Pool.Exec(ctx, `
		UPDATE workouts
		SET name = $2, day = $3, exercises = $4, updated_at = now()
		WHERE id = $1
	`, workoutID, name, day, exercisesJSON)

	return err
}

// UpdateWorkoutWithRestDay updates a workout including is_rest_day flag
func (db *DB) UpdateWorkoutWithRestDay(ctx context.Context, workoutID uuid.UUID, name, day string, exercises []Exercise, isRestDay bool) error {
	exercisesJSON, _ := json.Marshal(exercises)

	_, err := db.Pool.Exec(ctx, `
		UPDATE workouts
		SET name = $2, day = $3, exercises = $4, is_rest_day = $5, updated_at = now()
		WHERE id = $1
	`, workoutID, name, day, exercisesJSON, isRestDay)

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
		SELECT id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at, is_rest_day
		FROM workout_logs
		WHERE user_id = $1 AND date = $2
	`, userID, dateStr).Scan(
		&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesJSON, &cardioJSON,
		&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt, &wl.IsRestDay,
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

// SaveWorkoutLog creates or updates a workout log (without rest day - for backwards compatibility)
func (db *DB) SaveWorkoutLog(ctx context.Context, userID uuid.UUID, workoutName string, exercises []Exercise, cardio []Cardio, weight float64, date time.Time) (*WorkoutLog, error) {
	return db.SaveWorkoutLogWithRestDay(ctx, userID, workoutName, exercises, cardio, weight, date, false)
}

// SaveWorkoutLogWithRestDay creates or updates a workout log with rest day support
func (db *DB) SaveWorkoutLogWithRestDay(ctx context.Context, userID uuid.UUID, workoutName string, exercises []Exercise, cardio []Cardio, weight float64, date time.Time, isRestDay bool) (*WorkoutLog, error) {
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
			SET name = $2, completed_exercises = $3, cardio = $4, weight = $5, is_rest_day = $6, updated_at = now()
			WHERE id = $1
			RETURNING id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at, is_rest_day
		`, existingID, workoutName, exercisesJSON, cardioJSON, weight, isRestDay).Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesBytes, &cardioBytes,
			&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt, &wl.IsRestDay,
		)
	} else {
		// Create new
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO workout_logs (user_id, name, completed_exercises, cardio, weight, date, is_rest_day)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at, is_rest_day
		`, userID, workoutName, exercisesJSON, cardioJSON, weight, dateStr, isRestDay).Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesBytes, &cardioBytes,
			&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt, &wl.IsRestDay,
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
