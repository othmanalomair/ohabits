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

// MedicationWithDoses combines medication with its dose statuses for a day
type MedicationWithDoses struct {
	Medication
	DoseTaken []bool `json:"dose_taken"` // Status for each dose (indexed 0 to TimesPerDay-1)
}

// GetMedicationsForDay retrieves active medications for a specific day with dose statuses
// Excludes deleted medications
func (db *DB) GetMedicationsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]MedicationWithDoses, error) {
	dayName := weekdayToEnglish[date.Weekday()]
	dateStr := date.Format("2006-01-02")

	// First get all active medications
	rows, err := db.Pool.Query(ctx, `
		SELECT m.id, m.user_id, m.name, m.dosage, m.scheduled_days, m.times_per_day,
			   m.duration_type, m.start_date, m.end_date, COALESCE(m.notes, '') as notes,
			   COALESCE(m.icon, 'pill.fill') as icon, m.is_active,
			   m.created_at, m.updated_at
		FROM medications m
		WHERE m.user_id = $1 AND m.is_active = true AND COALESCE(m.is_deleted, false) = false
		ORDER BY m.created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medications []MedicationWithDoses
	for rows.Next() {
		var m MedicationWithDoses
		var daysJSON []byte
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysJSON, &m.TimesPerDay,
			&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.Icon, &m.IsActive,
			&m.CreatedAt, &m.UpdatedAt,
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
			// Initialize dose statuses
			m.DoseTaken = make([]bool, m.TimesPerDay)
			medications = append(medications, m)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Now get all medication logs for this date
	logRows, err := db.Pool.Query(ctx, `
		SELECT medication_id, dose_number, taken
		FROM medication_logs
		WHERE user_id = $1 AND date = $2
	`, userID, dateStr)
	if err != nil {
		return nil, err
	}
	defer logRows.Close()

	// Build a map of medication logs
	logMap := make(map[uuid.UUID]map[int]bool) // medication_id -> dose_number -> taken
	for logRows.Next() {
		var medID uuid.UUID
		var doseNum int
		var taken bool
		if err := logRows.Scan(&medID, &doseNum, &taken); err != nil {
			return nil, err
		}
		if logMap[medID] == nil {
			logMap[medID] = make(map[int]bool)
		}
		logMap[medID][doseNum] = taken
	}

	// Apply log statuses to medications
	for i := range medications {
		if doses, ok := logMap[medications[i].ID]; ok {
			for doseNum, taken := range doses {
				if doseNum >= 1 && doseNum <= len(medications[i].DoseTaken) {
					medications[i].DoseTaken[doseNum-1] = taken
				}
			}
		}
	}

	return medications, nil
}

// GetAllMedications retrieves all medications for a user (including deleted for sync)
func (db *DB) GetAllMedications(ctx context.Context, userID uuid.UUID) ([]Medication, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day,
			   duration_type, start_date, end_date, COALESCE(notes, '') as notes,
			   COALESCE(icon, 'pill.fill') as icon, is_active, COALESCE(is_deleted, false) as is_deleted,
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
			&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.Icon, &m.IsActive, &m.IsDeleted,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(daysJSON, &m.ScheduledDays)
		medications = append(medications, m)
	}

	return medications, rows.Err()
}

// GetActiveMedications retrieves only active (non-deleted) medications for management page
func (db *DB) GetActiveMedications(ctx context.Context, userID uuid.UUID) ([]Medication, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day,
			   duration_type, start_date, end_date, COALESCE(notes, '') as notes,
			   COALESCE(icon, 'pill.fill') as icon, is_active, COALESCE(is_deleted, false) as is_deleted,
			   created_at, updated_at
		FROM medications WHERE user_id = $1 AND COALESCE(is_deleted, false) = false
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
			&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.Icon, &m.IsActive, &m.IsDeleted,
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
func (db *DB) CreateMedication(ctx context.Context, userID uuid.UUID, name, dosage string, scheduledDays []string, timesPerDay int, durationType string, startDate, endDate *time.Time, notes, icon string) (*Medication, error) {
	daysJSON, _ := json.Marshal(scheduledDays)

	// Default icon if not provided
	if icon == "" {
		icon = "pill.fill"
	}

	var m Medication
	var daysBytes []byte
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO medications (user_id, name, dosage, scheduled_days, times_per_day, duration_type, start_date, end_date, notes, icon)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, name, dosage, scheduled_days, times_per_day, duration_type, start_date, end_date, COALESCE(notes, '') as notes, COALESCE(icon, 'pill.fill') as icon, is_active, created_at, updated_at
	`, userID, name, dosage, daysJSON, timesPerDay, durationType, startDate, endDate, notes, icon).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysBytes, &m.TimesPerDay,
		&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.Icon, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(daysBytes, &m.ScheduledDays)
	return &m, nil
}

// ToggleMedicationLog toggles the taken status of a specific dose for a medication on a date
func (db *DB) ToggleMedicationLog(ctx context.Context, userID, medicationID uuid.UUID, date time.Time, doseNumber int) (bool, error) {
	dateStr := date.Format("2006-01-02")

	// Check current status
	var taken bool
	err := db.Pool.QueryRow(ctx, `
		SELECT taken FROM medication_logs
		WHERE medication_id = $1 AND date = $2 AND dose_number = $3
	`, medicationID, dateStr, doseNumber).Scan(&taken)

	if err != nil {
		// No record exists, create one with taken = true
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO medication_logs (medication_id, user_id, taken, date, dose_number)
			VALUES ($1, $2, true, $3, $4)
		`, medicationID, userID, dateStr, doseNumber)
		return true, err
	}

	// Toggle existing record
	newStatus := !taken
	_, err = db.Pool.Exec(ctx, `
		UPDATE medication_logs SET taken = $4, updated_at = NOW()
		WHERE medication_id = $1 AND date = $2 AND dose_number = $3
	`, medicationID, dateStr, doseNumber, newStatus)

	return newStatus, err
}

// DeleteMedication soft-deletes a medication (for sync compatibility)
func (db *DB) DeleteMedication(ctx context.Context, medicationID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE medications SET is_deleted = true, updated_at = NOW() WHERE id = $1
	`, medicationID)
	return err
}

// HardDeleteMedication permanently deletes a medication (use sparingly)
func (db *DB) HardDeleteMedication(ctx context.Context, medicationID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM medications WHERE id = $1`, medicationID)
	return err
}

// UpdateMedication updates a medication
func (db *DB) UpdateMedication(ctx context.Context, medicationID uuid.UUID, name, dosage string, scheduledDays []string, timesPerDay int, durationType string, startDate, endDate *time.Time, notes, icon string, isActive bool) error {
	daysJSON, _ := json.Marshal(scheduledDays)

	// Default icon if not provided
	if icon == "" {
		icon = "pill.fill"
	}

	_, err := db.Pool.Exec(ctx, `
		UPDATE medications
		SET name = $2, dosage = $3, scheduled_days = $4, times_per_day = $5,
		    duration_type = $6, start_date = $7, end_date = $8, notes = $9,
		    icon = $10, is_active = $11, updated_at = now()
		WHERE id = $1
	`, medicationID, name, dosage, daysJSON, timesPerDay, durationType, startDate, endDate, notes, icon, isActive)

	return err
}

// GetMedicationByID retrieves a single medication by ID
func (db *DB) GetMedicationByID(ctx context.Context, medicationID uuid.UUID) (*Medication, error) {
	var m Medication
	var daysJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day,
			   duration_type, start_date, end_date, COALESCE(notes, '') as notes,
			   COALESCE(icon, 'pill.fill') as icon, is_active, COALESCE(is_deleted, false) as is_deleted,
			   created_at, updated_at
		FROM medications WHERE id = $1
	`, medicationID).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Dosage, &daysJSON, &m.TimesPerDay,
		&m.DurationType, &m.StartDate, &m.EndDate, &m.Notes, &m.Icon, &m.IsActive, &m.IsDeleted,
		&m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(daysJSON, &m.ScheduledDays)
	return &m, nil
}

// UpsertMedicationLog sets the medication log status directly (for sync)
func (db *DB) UpsertMedicationLog(ctx context.Context, userID, medicationID uuid.UUID, date time.Time, taken bool, doseNumber int) error {
	dateStr := date.Format("2006-01-02")

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO medication_logs (medication_id, user_id, taken, date, dose_number)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (medication_id, date, dose_number)
		DO UPDATE SET taken = $3, updated_at = NOW()
	`, medicationID, userID, taken, dateStr, doseNumber)

	return err
}

// GetMedicationDoseStatus gets the status of a specific dose (for HTMX updates)
func (db *DB) GetMedicationDoseStatus(ctx context.Context, medicationID uuid.UUID, date time.Time, doseNumber int) (bool, error) {
	dateStr := date.Format("2006-01-02")

	var taken bool
	err := db.Pool.QueryRow(ctx, `
		SELECT taken FROM medication_logs
		WHERE medication_id = $1 AND date = $2 AND dose_number = $3
	`, medicationID, dateStr, doseNumber).Scan(&taken)

	if err != nil {
		return false, nil // No record means not taken
	}
	return taken, nil
}
