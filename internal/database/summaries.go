package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetMonthlySummary retrieves the monthly summary for a specific month
func (db *DB) GetMonthlySummary(ctx context.Context, userID uuid.UUID, year, month int) (*MonthlySummary, error) {
	var s MonthlySummary
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, year, month, summary_text, is_ai_generated, created_at, updated_at
		FROM monthly_summaries
		WHERE user_id = $1 AND year = $2 AND month = $3
	`, userID, year, month).Scan(&s.ID, &s.UserID, &s.Year, &s.Month, &s.SummaryText, &s.IsAIGenerated, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No summary for this month
		}
		return nil, err
	}

	return &s, nil
}

// SaveMonthlySummary creates or updates a monthly summary (upsert)
func (db *DB) SaveMonthlySummary(ctx context.Context, userID uuid.UUID, year, month int, text string, isAI bool) (*MonthlySummary, error) {
	var s MonthlySummary
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO monthly_summaries (user_id, year, month, summary_text, is_ai_generated)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, year, month) DO UPDATE SET
			summary_text = $4,
			is_ai_generated = $5,
			updated_at = NOW()
		RETURNING id, user_id, year, month, summary_text, is_ai_generated, created_at, updated_at
	`, userID, year, month, text, isAI).Scan(&s.ID, &s.UserID, &s.Year, &s.Month, &s.SummaryText, &s.IsAIGenerated, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetAllNotesTextForMonth retrieves all notes text for a month, concatenated for AI processing
func (db *DB) GetAllNotesTextForMonth(ctx context.Context, userID uuid.UUID, year, month int) (string, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	rows, err := db.Pool.Query(ctx, `
		SELECT date, text
		FROM notes
		WHERE user_id = $1 AND date >= $2 AND date <= $3 AND text != ''
		ORDER BY date ASC
	`, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var date time.Time
		var text string
		if err := rows.Scan(&date, &text); err != nil {
			return "", err
		}
		// Format: "يوم 5: النص..."
		parts = append(parts, "يوم "+date.Format("2")+": "+text)
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return strings.Join(parts, "\n\n"), nil
}
