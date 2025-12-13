package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// GetCalendarEventsByUserID retrieves all calendar events for a user
func (db *DB) GetCalendarEventsByUserID(ctx context.Context, userID uuid.UUID) ([]CalendarEvent, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, event_type, event_date, end_date, is_recurring, notes, created_at, updated_at
		FROM calendar_events
		WHERE user_id = $1
		ORDER BY EXTRACT(MONTH FROM event_date), EXTRACT(DAY FROM event_date)
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		var notes *string
		if err := rows.Scan(&e.ID, &e.UserID, &e.Title, &e.EventType, &e.EventDate, &e.EndDate, &e.IsRecurring, &notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		if notes != nil {
			e.Notes = *notes
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// GetCalendarEventsByType retrieves calendar events of a specific type for a user
func (db *DB) GetCalendarEventsByType(ctx context.Context, userID uuid.UUID, eventType string) ([]CalendarEvent, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, event_type, event_date, end_date, is_recurring, notes, created_at, updated_at
		FROM calendar_events
		WHERE user_id = $1 AND event_type = $2
		ORDER BY EXTRACT(MONTH FROM event_date), EXTRACT(DAY FROM event_date)
	`, userID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		var notes *string
		if err := rows.Scan(&e.ID, &e.UserID, &e.Title, &e.EventType, &e.EventDate, &e.EndDate, &e.IsRecurring, &notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		if notes != nil {
			e.Notes = *notes
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// GetCalendarEventsForDay retrieves calendar events for a specific date (including recurring yearly events and date ranges)
func (db *DB) GetCalendarEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]CalendarEventForDay, error) {
	// For recurring events: match month and day (for start date or within range)
	// For non-recurring events: match exact date or within date range
	// Date range: event_date <= date <= end_date
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, event_type, event_date, end_date, is_recurring, notes, created_at, updated_at
		FROM calendar_events
		WHERE user_id = $1 AND (
			-- Single day events (no end_date)
			(end_date IS NULL AND (
				(is_recurring = true AND EXTRACT(MONTH FROM event_date) = $2 AND EXTRACT(DAY FROM event_date) = $3)
				OR
				(is_recurring = false AND event_date = $4)
			))
			OR
			-- Date range events (has end_date)
			(end_date IS NOT NULL AND (
				(is_recurring = false AND event_date <= $4 AND end_date >= $4)
				OR
				(is_recurring = true AND (
					-- For recurring ranges, we need to check if current date falls within the "adjusted" range for this year
					(EXTRACT(MONTH FROM event_date) < $2 OR (EXTRACT(MONTH FROM event_date) = $2 AND EXTRACT(DAY FROM event_date) <= $3))
					AND
					(EXTRACT(MONTH FROM end_date) > $2 OR (EXTRACT(MONTH FROM end_date) = $2 AND EXTRACT(DAY FROM end_date) >= $3))
				))
			))
		)
		ORDER BY event_type, title
	`, userID, int(date.Month()), date.Day(), date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []CalendarEventForDay
	for rows.Next() {
		var e CalendarEventForDay
		var notes *string
		if err := rows.Scan(&e.ID, &e.UserID, &e.Title, &e.EventType, &e.EventDate, &e.EndDate, &e.IsRecurring, &notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		if notes != nil {
			e.Notes = *notes
		}

		// Calculate years since the event (for birthdays)
		e.YearsAgo = date.Year() - e.EventDate.Year()
		e.IsToday = date.Month() == e.EventDate.Month() && date.Day() == e.EventDate.Day()

		events = append(events, e)
	}

	return events, rows.Err()
}

// GetDatesWithEventsForWeek returns a map of dates (YYYY-MM-DD format) that have events within the week
func (db *DB) GetDatesWithEventsForWeek(ctx context.Context, userID uuid.UUID, weekStart time.Time) (map[string]bool, error) {
	result := make(map[string]bool)

	// Check each day of the week
	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		dateStr := day.Format("2006-01-02")

		var count int
		err := db.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM calendar_events
			WHERE user_id = $1 AND (
				-- Single day events
				(end_date IS NULL AND (
					(is_recurring = true AND EXTRACT(MONTH FROM event_date) = $2 AND EXTRACT(DAY FROM event_date) = $3)
					OR
					(is_recurring = false AND event_date = $4)
				))
				OR
				-- Date range events
				(end_date IS NOT NULL AND (
					(is_recurring = false AND event_date <= $4 AND end_date >= $4)
					OR
					(is_recurring = true AND (
						(EXTRACT(MONTH FROM event_date) < $2 OR (EXTRACT(MONTH FROM event_date) = $2 AND EXTRACT(DAY FROM event_date) <= $3))
						AND
						(EXTRACT(MONTH FROM end_date) > $2 OR (EXTRACT(MONTH FROM end_date) = $2 AND EXTRACT(DAY FROM end_date) >= $3))
					))
				))
			)
		`, userID, int(day.Month()), day.Day(), dateStr).Scan(&count)
		if err != nil {
			return nil, err
		}

		if count > 0 {
			result[dateStr] = true
		}
	}

	return result, nil
}

// CreateCalendarEvent creates a new calendar event
func (db *DB) CreateCalendarEvent(ctx context.Context, userID uuid.UUID, title, eventType string, eventDate time.Time, endDate *time.Time, isRecurring bool, notes string) (*CalendarEvent, error) {
	var e CalendarEvent
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	var endDateStr *string
	if endDate != nil {
		s := endDate.Format("2006-01-02")
		endDateStr = &s
	}

	err := db.Pool.QueryRow(ctx, `
		INSERT INTO calendar_events (user_id, title, event_type, event_date, end_date, is_recurring, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, title, event_type, event_date, end_date, is_recurring, notes, created_at, updated_at
	`, userID, title, eventType, eventDate.Format("2006-01-02"), endDateStr, isRecurring, notesPtr).Scan(
		&e.ID, &e.UserID, &e.Title, &e.EventType, &e.EventDate, &e.EndDate, &e.IsRecurring, &notesPtr, &e.CreatedAt, &e.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if notesPtr != nil {
		e.Notes = *notesPtr
	}

	return &e, nil
}

// UpdateCalendarEvent updates an existing calendar event
func (db *DB) UpdateCalendarEvent(ctx context.Context, eventID uuid.UUID, title, eventType string, eventDate time.Time, endDate *time.Time, isRecurring bool, notes string) error {
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	var endDateStr *string
	if endDate != nil {
		s := endDate.Format("2006-01-02")
		endDateStr = &s
	}

	_, err := db.Pool.Exec(ctx, `
		UPDATE calendar_events
		SET title = $2, event_type = $3, event_date = $4, end_date = $5, is_recurring = $6, notes = $7, updated_at = NOW()
		WHERE id = $1
	`, eventID, title, eventType, eventDate.Format("2006-01-02"), endDateStr, isRecurring, notesPtr)

	return err
}

// DeleteCalendarEvent deletes a calendar event
func (db *DB) DeleteCalendarEvent(ctx context.Context, eventID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM calendar_events WHERE id = $1`, eventID)
	return err
}
