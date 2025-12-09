package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// weekdayToEnglish maps Go weekday to English day name (matching database format)
var weekdayToEnglish = map[time.Weekday]string{
	time.Sunday:    "Sunday",
	time.Monday:    "Monday",
	time.Tuesday:   "Tuesday",
	time.Wednesday: "Wednesday",
	time.Thursday:  "Thursday",
	time.Friday:    "Friday",
	time.Saturday:  "Saturday",
}

// GetMedicationsForDay retrieves active medications for a specific day with log status
func (db *DB) GetMedicationsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]MedicationWithLog, error) {
	dayName := weekdayToEnglish[date.Weekday()]
	dateStr := date.Format("2006-01-02")

	rows, err := db.Pool.Query(ctx, `
		SELECT m.id, m.user_id, m.name, m.dosage, m.scheduled_days, m.times_per_day,
			   m.duration_type, m.start_date, m.end_date, COALESCE(m.notes, '') as notes, m.is_active,
			   m.created_at, m.updated_at,
			   COALESCE(ml.taken, false) as taken
		FROM medications m
		LEFT JOIN medication_logs ml ON m.id = ml.medication_id AND ml.date = $2
		WHERE m.user_id = $1 AND m.is_active = true
		ORDER BY m.created_at
	`, userID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medications []MedicationWithLog
	for rows.Next() {
		var m MedicationWithLog
		var daysJSON []byte
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysJSON, &m.TimesPerDay,
			&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.IsActive,
			&m.CreatedAt, &m.UpdatedAt, &m.Taken,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(daysJSON, &m.ScheduledDays)

		// Check if medication is scheduled for this day
		isScheduled := len(m.ScheduledDays) == 0 // Empty means every day
		for _, day := range m.ScheduledDays {
			if day == dayName {
				isScheduled = true
				break
			}
		}

		// Check if within date range for limited duration
		if m.DurationType == "limited" {
			if m.StartDate != nil && date.Before(*m.StartDate) {
				isScheduled = false
			}
			if m.EndDate != nil && date.After(*m.EndDate) {
				isScheduled = false
			}
		}

		if isScheduled {
			medications = append(medications, m)
		}
	}

	return medications, rows.Err()
}

// GetAllMedications retrieves all medications for a user
func (db *DB) GetAllMedications(ctx context.Context, userID uuid.UUID) ([]Medication, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day,
			   duration_type, start_date, end_date, COALESCE(notes, '') as notes, is_active,
			   created_at, updated_at
		FROM medications WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medications []Medication
	for rows.Next() {
		var m Medication
		var daysJSON []byte
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysJSON, &m.TimesPerDay,
			&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.IsActive,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(daysJSON, &m.ScheduledDays)
		medications = append(medications, m)
	}

	return medications, rows.Err()
}

// CreateMedication creates a new medication
func (db *DB) CreateMedication(ctx context.Context, userID uuid.UUID, name, dosage string, scheduledDays []string, timesPerDay int, durationType string, startDate, endDate *time.Time, notes string) (*Medication, error) {
	daysJSON, _ := json.Marshal(scheduledDays)

	var m Medication
	var daysBytes []byte
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO medications (user_id, name, dosage, scheduled_days, times_per_day, duration_type, start_date, end_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, name, dosage, scheduled_days, times_per_day, duration_type, start_date, end_date, COALESCE(notes, '') as notes, is_active, created_at, updated_at
	`, userID, name, dosage, daysJSON, timesPerDay, durationType, startDate, endDate, notes).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysBytes, &m.TimesPerDay,
		&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(daysBytes, &m.ScheduledDays)
	return &m, nil
}

// ToggleMedicationLog toggles the taken status of a medication for a date
func (db *DB) ToggleMedicationLog(ctx context.Context, userID, medicationID uuid.UUID, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")

	// Check current status
	var taken bool
	err := db.Pool.QueryRow(ctx, `
		SELECT taken FROM medication_logs
		WHERE medication_id = $1 AND date = $2
	`, medicationID, dateStr).Scan(&taken)

	if err != nil {
		// No record exists, create one with taken = true
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO medication_logs (medication_id, user_id, taken, date)
			VALUES ($1, $2, true, $3)
		`, medicationID, userID, dateStr)
		return true, err
	}

	// Toggle existing record
	newStatus := !taken
	_, err = db.Pool.Exec(ctx, `
		UPDATE medication_logs SET taken = $3
		WHERE medication_id = $1 AND date = $2
	`, medicationID, dateStr, newStatus)

	return newStatus, err
}

// DeleteMedication deletes a medication
func (db *DB) DeleteMedication(ctx context.Context, medicationID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM medications WHERE id = $1`, medicationID)
	return err
}

// UpdateMedication updates a medication
func (db *DB) UpdateMedication(ctx context.Context, medicationID uuid.UUID, name, dosage string, scheduledDays []string, timesPerDay int, durationType string, startDate, endDate *time.Time, notes string, isActive bool) error {
	daysJSON, _ := json.Marshal(scheduledDays)

	_, err := db.Pool.Exec(ctx, `
		UPDATE medications
		SET name = $2, dosage = $3, scheduled_days = $4, times_per_day = $5,
		    duration_type = $6, start_date = $7, end_date = $8, notes = $9,
		    is_active = $10, updated_at = now()
		WHERE id = $1
	`, medicationID, name, dosage, daysJSON, timesPerDay, durationType, startDate, endDate, notes, isActive)

	return err
}

// GetMedicationByID retrieves a single medication by ID
func (db *DB) GetMedicationByID(ctx context.Context, medicationID uuid.UUID) (*Medication, error) {
	var m Medication
	var daysJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day,
			   duration_type, start_date, end_date, COALESCE(notes, '') as notes, is_active,
			   created_at, updated_at
		FROM medications WHERE id = $1
	`, medicationID).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysJSON, &m.TimesPerDay,
		&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(daysJSON, &m.ScheduledDays)
	return &m, nil
}
