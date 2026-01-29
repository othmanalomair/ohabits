package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SyncAllData contains all user data for full sync
type SyncAllData struct {
	Habits            []Habit            `json:"habits"`
	HabitCompletions  []HabitCompletion  `json:"habitCompletions"`
	Medications       []Medication       `json:"medications"`
	MedicationLogs    []MedicationLog    `json:"medicationLogs"`
	MoodRatings       []MoodRating       `json:"moodRatings"`
	DailyNotes        []Note             `json:"dailyNotes"`
	Todos             []Todo             `json:"todos"`
	Events            []CalendarEvent    `json:"events"`
	WorkoutTemplates  []Workout          `json:"workoutTemplates"`
	WorkoutLogs       []WorkoutLog       `json:"workoutLogs"`
	MarkdownNotes     []MarkdownNote     `json:"markdownNotes"`
	LastSyncTimestamp time.Time          `json:"lastSyncTimestamp"`
}

// SyncChangesData contains data changed since a specific timestamp
type SyncChangesData struct {
	Habits            []Habit            `json:"habits,omitempty"`
	HabitCompletions  []HabitCompletion  `json:"habitCompletions,omitempty"`
	Medications       []Medication       `json:"medications,omitempty"`
	MedicationLogs    []MedicationLog    `json:"medicationLogs,omitempty"`
	MoodRatings       []MoodRating       `json:"moodRatings,omitempty"`
	DailyNotes        []Note             `json:"dailyNotes,omitempty"`
	Todos             []Todo             `json:"todos,omitempty"`
	Events            []CalendarEvent    `json:"events,omitempty"`
	WorkoutTemplates  []Workout          `json:"workoutTemplates,omitempty"`
	WorkoutLogs       []WorkoutLog       `json:"workoutLogs,omitempty"`
	MarkdownNotes     []MarkdownNote     `json:"markdownNotes,omitempty"`
	LastSyncTimestamp time.Time          `json:"lastSyncTimestamp"`
}

// SyncPushItem represents a single item to be synced from the client
type SyncPushItem struct {
	LocalID   string          `json:"local_id"`
	ServerID  *string         `json:"server_id,omitempty"`
	Type      string          `json:"type"`
	IsDeleted bool            `json:"is_deleted"`
	UpdatedAt time.Time       `json:"updated_at"`
	Data      json.RawMessage `json:"data"`
}

// SyncPushResult contains the result of a push operation
type SyncPushResult struct {
	LocalID  string `json:"local_id"`
	ServerID string `json:"server_id"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// GetAllSyncData retrieves all user data for full sync
func (db *DB) GetAllSyncData(ctx context.Context, userID uuid.UUID) (*SyncAllData, error) {
	data := &SyncAllData{
		LastSyncTimestamp: time.Now(),
	}

	// Get habits
	habits, err := db.GetHabitsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.Habits = habits

	// Get habit completions
	completions, err := db.getAllHabitCompletions(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.HabitCompletions = completions

	// Get medications
	medications, err := db.GetAllMedications(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.Medications = medications

	// Get medication logs
	medLogs, err := db.getAllMedicationLogs(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.MedicationLogs = medLogs

	// Get mood ratings
	moods, err := db.getAllMoodRatings(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.MoodRatings = moods

	// Get daily notes
	notes, err := db.getAllNotes(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.DailyNotes = notes

	// Get todos
	todos, err := db.getAllTodos(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.Todos = todos

	// Get calendar events
	events, err := db.GetCalendarEventsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.Events = events

	// Get workout templates
	workouts, err := db.GetWorkoutsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.WorkoutTemplates = workouts

	// Get workout logs
	workoutLogs, err := db.getAllWorkoutLogs(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.WorkoutLogs = workoutLogs

	// Get markdown notes (blog)
	markdownNotes, err := db.GetAllBlogPostsIncludingDeleted(ctx, userID)
	if err != nil {
		return nil, err
	}
	data.MarkdownNotes = markdownNotes

	return data, nil
}

// GetSyncChangesSince retrieves all data changed since the given timestamp
func (db *DB) GetSyncChangesSince(ctx context.Context, userID uuid.UUID, since time.Time) (*SyncChangesData, error) {
	data := &SyncChangesData{
		LastSyncTimestamp: time.Now(),
	}

	// Get habits updated since timestamp
	habits, err := db.getHabitsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(habits) > 0 {
		data.Habits = habits
	}

	// Get habit completions updated since timestamp
	completions, err := db.getHabitCompletionsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(completions) > 0 {
		data.HabitCompletions = completions
	}

	// Get medications updated since timestamp
	medications, err := db.getMedicationsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(medications) > 0 {
		data.Medications = medications
	}

	// Get medication logs updated since timestamp
	medLogs, err := db.getMedicationLogsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(medLogs) > 0 {
		data.MedicationLogs = medLogs
	}

	// Get mood ratings created since timestamp
	moods, err := db.getMoodRatingsCreatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(moods) > 0 {
		data.MoodRatings = moods
	}

	// Get daily notes updated since timestamp
	notes, err := db.getNotesUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(notes) > 0 {
		data.DailyNotes = notes
	}

	// Get todos created since timestamp
	todos, err := db.getTodosCreatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(todos) > 0 {
		data.Todos = todos
	}

	// Get calendar events updated since timestamp
	events, err := db.getEventsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(events) > 0 {
		data.Events = events
	}

	// Get workout templates updated since timestamp
	workouts, err := db.getWorkoutsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(workouts) > 0 {
		data.WorkoutTemplates = workouts
	}

	// Get workout logs updated since timestamp
	workoutLogs, err := db.getWorkoutLogsUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(workoutLogs) > 0 {
		data.WorkoutLogs = workoutLogs
	}

	// Get markdown notes updated since timestamp
	markdownNotes, err := db.getMarkdownNotesUpdatedSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if len(markdownNotes) > 0 {
		data.MarkdownNotes = markdownNotes
	}

	return data, nil
}

// Helper functions for getting all data

func (db *DB) getAllHabitCompletions(ctx context.Context, userID uuid.UUID) ([]HabitCompletion, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, habit_id, user_id, completed, date, created_at
		FROM habits_completions WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var completions []HabitCompletion
	for rows.Next() {
		var c HabitCompletion
		if err := rows.Scan(&c.ID, &c.HabitID, &c.UserID, &c.Completed, &c.Date, &c.CreatedAt); err != nil {
			return nil, err
		}
		completions = append(completions, c)
	}

	return completions, rows.Err()
}

func (db *DB) getAllMedicationLogs(ctx context.Context, userID uuid.UUID) ([]MedicationLog, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, medication_id, user_id, taken, date, created_at
		FROM medication_logs WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []MedicationLog
	for rows.Next() {
		var l MedicationLog
		if err := rows.Scan(&l.ID, &l.MedicationID, &l.UserID, &l.Taken, &l.Date, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}

func (db *DB) getAllMoodRatings(ctx context.Context, userID uuid.UUID) ([]MoodRating, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, rating, date, created_at
		FROM mood_ratings WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []MoodRating
	for rows.Next() {
		var r MoodRating
		if err := rows.Scan(&r.ID, &r.UserID, &r.Rating, &r.Date, &r.CreatedAt); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}

	return ratings, rows.Err()
}

func (db *DB) getAllNotes(ctx context.Context, userID uuid.UUID) ([]Note, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, date, created_at, updated_at
		FROM notes WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Text, &n.Date, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	return notes, rows.Err()
}

func (db *DB) getAllTodos(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, completed, date, created_at, false as is_overdue, false as is_deleted
		FROM todos WHERE user_id = $1 AND is_deleted = false
		ORDER BY date DESC, created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Text, &t.Completed, &t.Date, &t.CreatedAt, &t.IsOverdue, &t.IsDeleted); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}

	return todos, rows.Err()
}

func (db *DB) getAllWorkoutLogs(ctx context.Context, userID uuid.UUID) ([]WorkoutLog, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at
		FROM workout_logs WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []WorkoutLog
	for rows.Next() {
		var wl WorkoutLog
		var exercisesJSON, cardioJSON []byte
		if err := rows.Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesJSON, &cardioJSON,
			&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(exercisesJSON, &wl.CompletedExercises)
		if cardioJSON != nil {
			json.Unmarshal(cardioJSON, &wl.Cardio)
		}
		logs = append(logs, wl)
	}

	return logs, rows.Err()
}

// Helper functions for getting data since timestamp

func (db *DB) getHabitsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Habit, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, scheduled_days, created_at, updated_at
		FROM habits WHERE user_id = $1 AND updated_at > $2
		ORDER BY created_at
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []Habit
	for rows.Next() {
		var h Habit
		var daysJSON []byte
		if err := rows.Scan(&h.ID, &h.UserID, &h.Name, &daysJSON, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(daysJSON, &h.ScheduledDays)
		habits = append(habits, h)
	}

	return habits, rows.Err()
}

func (db *DB) getHabitCompletionsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]HabitCompletion, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, habit_id, user_id, completed, date, created_at
		FROM habits_completions WHERE user_id = $1 AND updated_at > $2
		ORDER BY date DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var completions []HabitCompletion
	for rows.Next() {
		var c HabitCompletion
		if err := rows.Scan(&c.ID, &c.HabitID, &c.UserID, &c.Completed, &c.Date, &c.CreatedAt); err != nil {
			return nil, err
		}
		completions = append(completions, c)
	}

	return completions, rows.Err()
}

func (db *DB) getMedicationsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Medication, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day,
			   duration_type, start_date, end_date, COALESCE(notes, '') as notes, is_active,
			   created_at, updated_at
		FROM medications WHERE user_id = $1 AND updated_at > $2
		ORDER BY created_at
	`, userID, since)
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

func (db *DB) getMedicationLogsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]MedicationLog, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, medication_id, user_id, taken, date, created_at
		FROM medication_logs WHERE user_id = $1 AND updated_at > $2
		ORDER BY date DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []MedicationLog
	for rows.Next() {
		var l MedicationLog
		if err := rows.Scan(&l.ID, &l.MedicationID, &l.UserID, &l.Taken, &l.Date, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}

func (db *DB) getMoodRatingsCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]MoodRating, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, rating, date, created_at
		FROM mood_ratings WHERE user_id = $1 AND created_at > $2
		ORDER BY date DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []MoodRating
	for rows.Next() {
		var r MoodRating
		if err := rows.Scan(&r.ID, &r.UserID, &r.Rating, &r.Date, &r.CreatedAt); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}

	return ratings, rows.Err()
}

func (db *DB) getNotesUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Note, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, date, created_at, updated_at
		FROM notes WHERE user_id = $1 AND updated_at > $2
		ORDER BY date DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Text, &n.Date, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	return notes, rows.Err()
}

func (db *DB) getTodosCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Todo, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, text, completed, date, created_at, false as is_overdue, is_deleted
		FROM todos WHERE user_id = $1 AND updated_at > $2
		ORDER BY date DESC, created_at ASC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Text, &t.Completed, &t.Date, &t.CreatedAt, &t.IsOverdue, &t.IsDeleted); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}

	return todos, rows.Err()
}

func (db *DB) getEventsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]CalendarEvent, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, event_type, event_date, end_date, is_recurring, notes, created_at, updated_at
		FROM calendar_events
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY event_date
	`, userID, since)
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

func (db *DB) getWorkoutsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Workout, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at
		FROM workouts
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY display_order
	`, userID, since)
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

func (db *DB) getWorkoutLogsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]WorkoutLog, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at
		FROM workout_logs WHERE user_id = $1 AND updated_at > $2
		ORDER BY date DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []WorkoutLog
	for rows.Next() {
		var wl WorkoutLog
		var exercisesJSON, cardioJSON []byte
		if err := rows.Scan(
			&wl.ID, &wl.UserID, &wl.WorkoutName, &exercisesJSON, &cardioJSON,
			&wl.Weight, &wl.Date, &wl.CreatedAt, &wl.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(exercisesJSON, &wl.CompletedExercises)
		if cardioJSON != nil {
			json.Unmarshal(cardioJSON, &wl.Cardio)
		}
		logs = append(logs, wl)
	}

	return logs, rows.Err()
}

func (db *DB) getMarkdownNotesUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]MarkdownNote, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, content, is_rtl, is_deleted, created_at, updated_at
		FROM markdown_notes
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []MarkdownNote
	for rows.Next() {
		var p MarkdownNote
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.IsRTL, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

// Sync push operations - handle creates, updates, and deletes

// SyncPushHabit handles syncing a habit from the client
func (db *DB) SyncPushHabit(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		if serverID == nil {
			// Todo was deleted before ever syncing - nothing to do on server
			return "", nil
		}
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.DeleteHabit(ctx, id)
	}

	var habitData struct {
		Name          string   `json:"name"`
		ScheduledDays []string `json:"scheduled_days"`
	}
	if err := json.Unmarshal(data, &habitData); err != nil {
		return "", err
	}

	if serverID != nil {
		// Update existing
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.UpdateHabit(ctx, id, habitData.Name, habitData.ScheduledDays)
	}

	// Create new
	habit, err := db.CreateHabit(ctx, userID, habitData.Name, habitData.ScheduledDays)
	if err != nil {
		return "", err
	}
	return habit.ID.String(), nil
}

// SyncPushMedication handles syncing a medication from the client
func (db *DB) SyncPushMedication(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		if serverID == nil {
			// Todo was deleted before ever syncing - nothing to do on server
			return "", nil
		}
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.DeleteMedication(ctx, id)
	}

	var medData struct {
		Name          string     `json:"name"`
		Dosage        string     `json:"dosage"`
		ScheduledDays []string   `json:"scheduled_days"`
		TimesPerDay   int        `json:"times_per_day"`
		DurationType  string     `json:"duration_type"`
		StartDate     *time.Time `json:"start_date"`
		EndDate       *time.Time `json:"end_date"`
		Notes         string     `json:"notes"`
		IsActive      bool       `json:"is_active"`
	}
	if err := json.Unmarshal(data, &medData); err != nil {
		return "", err
	}

	if serverID != nil {
		// Update existing
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.UpdateMedication(ctx, id, medData.Name, medData.Dosage, medData.ScheduledDays, medData.TimesPerDay, medData.DurationType, medData.StartDate, medData.EndDate, medData.Notes, medData.IsActive)
	}

	// Create new
	med, err := db.CreateMedication(ctx, userID, medData.Name, medData.Dosage, medData.ScheduledDays, medData.TimesPerDay, medData.DurationType, medData.StartDate, medData.EndDate, medData.Notes)
	if err != nil {
		return "", err
	}
	return med.ID.String(), nil
}

// SyncPushTodo handles syncing a todo from the client
func (db *DB) SyncPushTodo(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		if serverID == nil {
			// Todo was deleted before ever syncing - nothing to do on server
			return "", nil
		}
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.DeleteTodo(ctx, id)
	}

	var todoData struct {
		Text      string    `json:"text"`
		Completed bool      `json:"completed"`
		Date      time.Time `json:"date"`
	}
	if err := json.Unmarshal(data, &todoData); err != nil {
		return "", err
	}

	if serverID != nil {
		// Update existing - just update the text
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		if err := db.UpdateTodoWithCompleted(ctx, id, todoData.Text, todoData.Completed); err != nil {
			return "", err
		}
		// Also toggle if needed
		return *serverID, nil
	}

	// Create new
	todo, err := db.CreateTodo(ctx, userID, todoData.Text, todoData.Date)
	if err != nil {
		return "", err
	}
	return todo.ID.String(), nil
}

// SyncPushNote handles syncing a daily note from the client
func (db *DB) SyncPushNote(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	var noteData struct {
		Text string    `json:"text"`
		Date time.Time `json:"date"`
	}
	if err := json.Unmarshal(data, &noteData); err != nil {
		return "", err
	}

	// Notes use upsert based on date, so we always use SaveNote
	note, err := db.SaveNote(ctx, userID, noteData.Text, noteData.Date)
	if err != nil {
		return "", err
	}
	return note.ID.String(), nil
}

// SyncPushMood handles syncing a mood rating from the client
func (db *DB) SyncPushMood(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	var moodData struct {
		Rating int       `json:"rating"`
		Date   time.Time `json:"date"`
	}
	if err := json.Unmarshal(data, &moodData); err != nil {
		return "", err
	}

	// Moods use upsert based on date, so we always use SaveMood
	mood, err := db.SaveMood(ctx, userID, moodData.Rating, moodData.Date)
	if err != nil {
		return "", err
	}
	return mood.ID.String(), nil
}

// SyncPushEvent handles syncing a calendar event from the client
func (db *DB) SyncPushEvent(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		if serverID == nil {
			// Todo was deleted before ever syncing - nothing to do on server
			return "", nil
		}
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.DeleteCalendarEvent(ctx, id)
	}

	var eventData struct {
		Title       string     `json:"title"`
		EventType   string     `json:"event_type"`
		EventDate   time.Time  `json:"event_date"`
		EndDate     *time.Time `json:"end_date"`
		IsRecurring bool       `json:"is_recurring"`
		Notes       string     `json:"notes"`
	}
	if err := json.Unmarshal(data, &eventData); err != nil {
		return "", err
	}

	if serverID != nil {
		// Update existing
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.UpdateCalendarEvent(ctx, id, eventData.Title, eventData.EventType, eventData.EventDate, eventData.EndDate, eventData.IsRecurring, eventData.Notes)
	}

	// Create new
	event, err := db.CreateCalendarEvent(ctx, userID, eventData.Title, eventData.EventType, eventData.EventDate, eventData.EndDate, eventData.IsRecurring, eventData.Notes)
	if err != nil {
		return "", err
	}
	return event.ID.String(), nil
}

// SyncPushWorkout handles syncing a workout template from the client
func (db *DB) SyncPushWorkout(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		if serverID == nil {
			// Todo was deleted before ever syncing - nothing to do on server
			return "", nil
		}
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.DeleteWorkout(ctx, id)
	}

	var workoutData struct {
		Name      string     `json:"name"`
		Day       string     `json:"day"`
		Exercises []Exercise `json:"exercises"`
	}
	if err := json.Unmarshal(data, &workoutData); err != nil {
		return "", err
	}

	if serverID != nil {
		// Update existing
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.UpdateWorkout(ctx, id, workoutData.Name, workoutData.Day, workoutData.Exercises)
	}

	// Create new
	workout, err := db.CreateWorkout(ctx, userID, workoutData.Name, workoutData.Day, workoutData.Exercises)
	if err != nil {
		return "", err
	}
	return workout.ID.String(), nil
}

// SyncPushWorkoutLog handles syncing a workout log from the client
func (db *DB) SyncPushWorkoutLog(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	var logData struct {
		WorkoutName string     `json:"workout_name"`
		Exercises   []Exercise `json:"completed_exercises"`
		Cardio      []Cardio   `json:"cardio"`
		Weight      float64    `json:"weight"`
		Date        time.Time  `json:"date"`
	}
	if err := json.Unmarshal(data, &logData); err != nil {
		return "", err
	}

	// Workout logs use upsert based on date
	log, err := db.SaveWorkoutLog(ctx, userID, logData.WorkoutName, logData.Exercises, logData.Cardio, logData.Weight, logData.Date)
	if err != nil {
		return "", err
	}
	return log.ID.String(), nil
}

// SyncPushMarkdownNote handles syncing a blog post from the client
func (db *DB) SyncPushMarkdownNote(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		if serverID == nil {
			// Todo was deleted before ever syncing - nothing to do on server
			return "", nil
		}
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		return *serverID, db.DeleteBlogPost(ctx, userID, id)
	}

	var noteData struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &noteData); err != nil {
		return "", err
	}

	if serverID != nil {
		// Update existing
		id, err := uuid.Parse(*serverID)
		if err != nil {
			return "", err
		}
		note, err := db.UpdateBlogPost(ctx, userID, id, noteData.Title, noteData.Content)
		if err != nil {
			return "", err
		}
		return note.ID.String(), nil
	}

	// Create new
	note, err := db.CreateBlogPost(ctx, userID, noteData.Title)
	if err != nil {
		return "", err
	}
	// Update content if provided
	if noteData.Content != "" {
		note, err = db.UpdateBlogPost(ctx, userID, note.ID, noteData.Title, noteData.Content)
		if err != nil {
			return "", err
		}
	}
	return note.ID.String(), nil
}

// SyncPushHabitCompletion handles syncing a habit completion from the client
func (db *DB) SyncPushHabitCompletion(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	var compData struct {
		HabitID   string    `json:"habit_id"`
		Completed bool      `json:"completed"`
		Date      time.Time `json:"date"`
	}
	if err := json.Unmarshal(data, &compData); err != nil {
		return "", err
	}

	habitID, err := uuid.Parse(compData.HabitID)
	if err != nil {
		return "", err
	}

	// Use upsert to set the exact completion status
	err = db.UpsertHabitCompletion(ctx, userID, habitID, compData.Date, compData.Completed)
	if err != nil {
		return "", err
	}

	// Return a composite ID
	return habitID.String() + "_" + compData.Date.Format("2006-01-02"), nil
}

// SyncPushMedicationLog handles syncing a medication log from the client
func (db *DB) SyncPushMedicationLog(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	var logData struct {
		MedicationID string    `json:"medication_id"`
		Taken        bool      `json:"taken"`
		Date         time.Time `json:"date"`
	}
	if err := json.Unmarshal(data, &logData); err != nil {
		return "", err
	}

	medID, err := uuid.Parse(logData.MedicationID)
	if err != nil {
		return "", err
	}

	// Use upsert to set the exact completion status
	err = db.UpsertMedicationLog(ctx, userID, medID, logData.Date, logData.Taken)
	if err != nil {
		return "", err
	}

	// Return a composite ID
	return medID.String() + "_" + logData.Date.Format("2006-01-02"), nil
}
