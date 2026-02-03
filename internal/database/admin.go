package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AdminStats holds general statistics for admin dashboard
type AdminStats struct {
	TotalUsers      int `json:"total_users"`
	ActiveThisWeek  int `json:"active_this_week"`
	AdminCount      int `json:"admin_count"`
	SubscribedCount int `json:"subscribed_count"`
	NormalCount     int `json:"normal_count"`
}

// UserStats holds statistics for a single user
type UserStats struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	AvatarURL    *string    `json:"avatar_url"`
	Role         int        `json:"role"`
	HabitsCount  int        `json:"habits_count"`
	NotesCount   int        `json:"notes_count"`
	BlogCount    int        `json:"blog_count"`
	LastActivity *time.Time `json:"last_activity"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GetAdminStats returns overall system statistics
func (db *DB) GetAdminStats(ctx context.Context) (*AdminStats, error) {
	stats := &AdminStats{}

	// Total users
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Active users this week (users with habit completions, notes, or mood ratings in last 7 days)
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM habits_completions WHERE created_at > NOW() - INTERVAL '7 days'
			UNION
			SELECT user_id FROM notes WHERE created_at > NOW() - INTERVAL '7 days'
			UNION
			SELECT user_id FROM mood_ratings WHERE created_at > NOW() - INTERVAL '7 days'
		) AS active_users
	`).Scan(&stats.ActiveThisWeek)
	if err != nil {
		return nil, err
	}

	// Count by role
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 1`).Scan(&stats.AdminCount)
	if err != nil {
		return nil, err
	}

	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 3`).Scan(&stats.SubscribedCount)
	if err != nil {
		return nil, err
	}

	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 2`).Scan(&stats.NormalCount)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetAllUsersStats returns statistics for all users
func (db *DB) GetAllUsersStats(ctx context.Context) ([]UserStats, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT
			u.id,
			u.email,
			u.display_name,
			u.avatar_url,
			u.role,
			u.created_at,
			COALESCE(h.habits_count, 0) as habits_count,
			COALESCE(n.notes_count, 0) as notes_count,
			COALESCE(b.blog_count, 0) as blog_count,
			GREATEST(
				COALESCE(hc.last_habit, u.created_at),
				COALESCE(nc.last_note, u.created_at),
				COALESCE(mc.last_mood, u.created_at)
			) as last_activity
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) as habits_count
			FROM habits WHERE is_deleted = false
			GROUP BY user_id
		) h ON u.id = h.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as notes_count
			FROM notes
			GROUP BY user_id
		) n ON u.id = n.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as blog_count
			FROM markdown_notes WHERE is_deleted = false
			GROUP BY user_id
		) b ON u.id = b.user_id
		LEFT JOIN (
			SELECT user_id, MAX(created_at) as last_habit
			FROM habits_completions
			GROUP BY user_id
		) hc ON u.id = hc.user_id
		LEFT JOIN (
			SELECT user_id, MAX(created_at) as last_note
			FROM notes
			GROUP BY user_id
		) nc ON u.id = nc.user_id
		LEFT JOIN (
			SELECT user_id, MAX(created_at) as last_mood
			FROM mood_ratings
			GROUP BY user_id
		) mc ON u.id = mc.user_id
		ORDER BY last_activity DESC NULLS LAST
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserStats
	for rows.Next() {
		var u UserStats
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.DisplayName,
			&u.AvatarURL,
			&u.Role,
			&u.CreatedAt,
			&u.HabitsCount,
			&u.NotesCount,
			&u.BlogCount,
			&u.LastActivity,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// UpdateUserRole updates a user's role
func (db *DB) UpdateUserRole(ctx context.Context, userID uuid.UUID, role int) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE users SET role = $2, updated_at = NOW()
		WHERE id = $1
	`, userID, role)
	return err
}

// DeleteUser deletes a user and all their associated data
func (db *DB) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Use a transaction to ensure all deletes succeed or none do
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete in order of dependencies (children first, then parents)
	// Only tables that have user_id column
	tables := []string{
		"habits_completions",
		"habits",
		"medication_logs",
		"medications",
		"todos",
		"notes",
		"mood_ratings",
		"daily_images",
		"workout_logs",
		"workouts",
		"blog_images",
		"markdown_notes",
		"calendar_events",
		"user_settings",
		"monthly_summaries",
		"episode_tracking",
		"episodes",
		"shows",
		"task_attachments",
		"task_comments",
		"tasks",
		"projects",
	}

	for _, table := range tables {
		_, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE user_id = $1", userID)
		if err != nil {
			return err
		}
	}

	// Finally delete the user
	_, err = tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
