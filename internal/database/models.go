package database

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account
type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`          // Never expose password
	AppleID     *string   `json:"apple_id"`   // Apple Sign-In identifier
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"` // nullable
	Role        int       `json:"role"`       // 1=admin, 2=normal, 3=subscribed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetAvatarURL returns the avatar URL or empty string if nil
func (u *User) GetAvatarURL() string {
	if u.AvatarURL != nil {
		return *u.AvatarURL
	}
	return ""
}

// Habit represents a habit to track
type Habit struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Name          string    `json:"name"`
	Icon          string    `json:"icon"`
	ScheduledDays []string  `json:"scheduled_days"` // Day names: "Sunday", "Monday", etc.
	CreatedAt     time.Time  `json:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	IsDeleted     bool       `json:"is_deleted"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IsScheduledFor checks if habit is scheduled for a given weekday
func (h *Habit) IsScheduledFor(weekday time.Weekday) bool {
	dayName := weekdayToName(weekday)
	for _, day := range h.ScheduledDays {
		if day == dayName {
			return true
		}
	}
	return false
}

// weekdayToName converts time.Weekday to English day name
func weekdayToName(w time.Weekday) string {
	names := map[time.Weekday]string{
		time.Sunday:    "Sunday",
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
	}
	return names[w]
}

// HabitCompletion tracks daily habit completion
type HabitCompletion struct {
	ID        uuid.UUID `json:"id"`
	HabitID   uuid.UUID `json:"habit_id"`
	UserID    uuid.UUID `json:"user_id"`
	Completed bool      `json:"completed"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

// HabitWithCompletion combines habit with its completion status
type HabitWithCompletion struct {
	Habit
	Completed bool `json:"completed"`
}

// Medication represents a medication to track
type Medication struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Name          string     `json:"name"`
	Dosage        string     `json:"dosage"`
	ScheduledDays []string   `json:"scheduled_days"` // Day names: "Sunday", "Monday", etc.
	TimesPerDay   int        `json:"times_per_day"`
	DurationType  string     `json:"duration_type"` // "lifetime" or "limited"
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	Notes         string     `json:"notes"`
	Icon          string     `json:"icon"`
	IsActive      bool       `json:"is_active"`
	IsDeleted     bool       `json:"is_deleted"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// MedicationLog tracks medication intake
type MedicationLog struct {
	ID           uuid.UUID `json:"id"`
	MedicationID uuid.UUID `json:"medication_id"`
	UserID       uuid.UUID `json:"user_id"`
	Taken        bool      `json:"taken"`
	DoseNumber   int       `json:"dose_number"`
	Date         time.Time `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
}

// MedicationWithLog combines medication with its log status
type MedicationWithLog struct {
	Medication
	Taken bool `json:"taken"`
}

// Todo represents a daily task
type Todo struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	IsOverdue bool      `json:"is_overdue"` // true if task is from a past date
	IsDeleted bool      `json:"is_deleted"` // true if todo was deleted
}

// Note represents a daily quick note
type Note struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MoodRating represents daily mood (1-5)
type MoodRating struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int       `json:"rating"` // 1-5
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

// DailyImage represents an image uploaded for a specific day
type DailyImage struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Date          time.Time `json:"date"`
	OriginalPath  string    `json:"original_path"`
	ThumbnailPath string    `json:"thumbnail_path"`
	Filename      string    `json:"filename"`
	MimeType      string    `json:"mime_type"`
	SizeBytes     int       `json:"size_bytes"`
	CreatedAt     time.Time  `json:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	IsDeleted     bool       `json:"is_deleted"`
}

// Workout represents a workout plan
type Workout struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Name         string     `json:"name"`
	Day          string     `json:"day"`
	Exercises    []Exercise `json:"exercises"`
	DisplayOrder int        `json:"display_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	IsRestDay    bool       `json:"is_rest_day"`
}

// Exercise represents a single exercise in a workout
type Exercise struct {
	Order int    `json:"order"`
	Name  string `json:"name"`
}

// WorkoutLog represents a completed workout session
type WorkoutLog struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	WorkoutName        string     `json:"workout_name"` // maps to "name" column in DB
	CompletedExercises []Exercise `json:"completed_exercises"`
	Cardio             []Cardio   `json:"cardio"` // array in DB
	Weight             float64    `json:"weight"`
	Date               time.Time  `json:"date"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	IsRestDay          bool       `json:"is_rest_day"`
}

// Cardio represents cardio activity
type Cardio struct {
	Name    string `json:"name"`
	Minutes int    `json:"minutes"`
}

// MarkdownNote represents a long-form note
type MarkdownNote struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsRTL     bool      `json:"is_rtl"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


// Show represents a TV show or anime
type Show struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ExternalID string    `json:"external_id"`
	ShowType   string    `json:"show_type"` // "tv" or "anime"
	Name       string    `json:"name"`
	Summary    string    `json:"summary"`
	Status     string    `json:"status"`
	ImageURL   string    `json:"image_url"`
	Premiered  time.Time `json:"premiered"`
	Ended      time.Time `json:"ended"`
	Network    string    `json:"network"`
	Genres     []string  `json:"genres"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Episode represents a show episode
type Episode struct {
	ID         uuid.UUID `json:"id"`
	ShowID     uuid.UUID `json:"show_id"`
	UserID     uuid.UUID `json:"user_id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Season     int       `json:"season"`
	Number     int       `json:"number"`
	Summary    string    `json:"summary"`
	AirDate    time.Time `json:"air_date"`
	ImageURL   string    `json:"image_url"`
	Watched    bool      `json:"watched"`
	UserRating *int      `json:"user_rating"`
	CreatedAt  time.Time `json:"created_at"`
}

// Project represents a project
type Project struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "active", "completed", "archived"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Task represents a project task
type Task struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	ProjectID    *uuid.UUID `json:"project_id"`
	ParentTaskID *uuid.UUID `json:"parent_task_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`   // "pending", "in_progress", "completed"
	Priority     string     `json:"priority"` // "low", "medium", "high"
	DueDate      *time.Time `json:"due_date"`
	DisplayOrder int        `json:"display_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CalendarEvent represents a calendar event (birthday, travel, holiday, anniversary, general)
type CalendarEvent struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	EventType   string     `json:"event_type"`   // birthday, travel, holiday, anniversary, general
	EventDate   time.Time  `json:"event_date"`   // Start date of the event
	EndDate     *time.Time `json:"end_date"`     // End date (optional, for date ranges)
	IsRecurring bool       `json:"is_recurring"` // Repeats yearly?
	Notes       string     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	IsDeleted   bool       `json:"is_deleted"`
}

// HasDateRange returns true if the event spans multiple days
func (e *CalendarEvent) HasDateRange() bool {
	return e.EndDate != nil && !e.EndDate.IsZero()
}

// CalendarEventForDay includes additional display info
type CalendarEventForDay struct {
	CalendarEvent
	YearsAgo int  `json:"years_ago"` // Years since original event (for birthdays)
	IsToday  bool `json:"is_today"`  // Is the event today?
}

// DashboardData holds all data for the main dashboard
type DashboardData struct {
	Date           time.Time
	Habits         []HabitWithCompletion
	Medications    []MedicationWithDoses
	Todos          []Todo
	Note           *Note
	Images         []DailyImage
	MoodRating     *MoodRating
	Workouts       []Workout
	WorkoutLog     *WorkoutLog
	CalendarEvents []CalendarEventForDay // Calendar events for the day
	WeekEvents     map[string]bool       // Dates with events in the current week
}

// DailyNoteEntry holds data for a single day in the daily notes page
type DailyNoteEntry struct {
	Date   time.Time
	Note   *Note
	Todos  []Todo
	Images []DailyImage
	Mood   *MoodRating
}

// MonthlySummary represents a monthly summary (AI-generated or manual)
type MonthlySummary struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Year          int       `json:"year"`
	Month         int       `json:"month"`
	SummaryText   string    `json:"summary_text"`
	IsAIGenerated bool      `json:"is_ai_generated"`
	CreatedAt     time.Time  `json:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	IsDeleted     bool       `json:"is_deleted"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserSettings represents user-specific settings like section visibility and ordering
type UserSettings struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	SectionConfigs []SectionConfig `json:"section_configs"`
	IsDeleted      bool            `json:"is_deleted"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// SectionConfig represents visibility and order settings for a homepage section
type SectionConfig struct {
	ID        string `json:"id"`
	IsVisible bool   `json:"is_visible"`
	Order     int    `json:"order"`
}

// BlogImage represents an image in a markdown note (blog post)
type BlogImage struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	MarkdownNoteID uuid.UUID  `json:"markdown_note_id"`
	OriginalPath   string     `json:"original_path"`
	ThumbnailPath  *string    `json:"thumbnail_path,omitempty"`
	Filename       string     `json:"filename"`
	MimeType       string     `json:"mime_type"`
	SizeBytes      int        `json:"size_bytes"`
	PositionMarker string     `json:"position_marker"`
	IsDeleted      bool       `json:"is_deleted"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
