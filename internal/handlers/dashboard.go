package handlers

import (
	"net/http"
	"time"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/labstack/echo/v4"
)

// Kuwait timezone (UTC+3)
var KuwaitTZ = time.FixedZone("Asia/Kuwait", 3*60*60)

// GetKuwaitTime returns current time in Kuwait timezone
func GetKuwaitTime() time.Time {
	return time.Now().In(KuwaitTZ)
}

// GetKuwaitDate returns date in Kuwait timezone (time set to midnight)
func GetKuwaitDate(t time.Time) time.Time {
	t = t.In(KuwaitTZ)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, KuwaitTZ)
}

// Dashboard renders the main dashboard page
func (h *Handler) Dashboard(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get date from query or use today (Kuwait time)
	dateStr := c.QueryParam("date")
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.ParseInLocation("2006-01-02", dateStr, KuwaitTZ)
		if err != nil {
			date = GetKuwaitDate(time.Now())
		}
	} else {
		date = GetKuwaitDate(time.Now())
	}

	// Load user
	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Load all dashboard data
	ctx := c.Request().Context()
	data := database.DashboardData{Date: date}

	// Habits for this day
	data.Habits, _ = h.DB.GetHabitsForDay(ctx, userID, date)

	// Medications for this day
	data.Medications, _ = h.DB.GetMedicationsForDay(ctx, userID, date)

	// Todos for this day
	data.Todos, _ = h.DB.GetTodosForDay(ctx, userID, date)

	// Note for this day
	data.Note, _ = h.DB.GetNoteForDay(ctx, userID, date)

	// Images for this day
	data.Images, _ = h.DB.GetImagesForDay(ctx, userID, date)

	// Mood for this day
	data.MoodRating, _ = h.DB.GetMoodForDay(ctx, userID, date)

	// Workouts
	dayName := getDayName(date.Weekday())
	data.Workouts, _ = h.DB.GetWorkoutsForDay(ctx, userID, dayName)

	// Workout log for this day
	data.WorkoutLog, _ = h.DB.GetWorkoutLogForDay(ctx, userID, date)

	return Render(c, http.StatusOK, pages.Dashboard(user, data))
}

// getDayName returns Arabic day name
func getDayName(weekday time.Weekday) string {
	days := map[time.Weekday]string{
		time.Sunday:    "الأحد",
		time.Monday:    "الاثنين",
		time.Tuesday:   "الثلاثاء",
		time.Wednesday: "الأربعاء",
		time.Thursday:  "الخميس",
		time.Friday:    "الجمعة",
		time.Saturday:  "السبت",
	}
	return days[weekday]
}
