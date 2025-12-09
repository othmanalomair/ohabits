package handlers

import (
	"fmt"
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
	var medErr error
	data.Medications, medErr = h.DB.GetMedicationsForDay(ctx, userID, date)
	if medErr != nil {
		println("❌ خطأ في جلب الأدوية:", medErr.Error())
	}
	println("📅 التاريخ:", date.Format("2006-01-02"), "اليوم:", date.Weekday().String())
	println("👤 UserID:", userID.String())
	println("💊 عدد الأدوية:", len(data.Medications))

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

// DailyNotesPage renders the daily notes page with monthly view
func (h *Handler) DailyNotesPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get user
	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get year and month from query params or use current
	now := GetKuwaitTime()
	year := now.Year()
	month := int(now.Month())

	if yearStr := c.QueryParam("year"); yearStr != "" {
		if y, err := parseIntParam(yearStr); err == nil && y >= 2020 && y <= 2100 {
			year = y
		}
	}

	if monthStr := c.QueryParam("month"); monthStr != "" {
		if m, err := parseIntParam(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	ctx := c.Request().Context()

	// Get all dates with content (notes, images, or todos)
	dates, err := h.DB.GetDatesWithContentForMonth(ctx, userID, year, month)
	if err != nil {
		dates = []time.Time{}
	}

	// Build entries for each date
	entries := make([]database.DailyNoteEntry, 0, len(dates))
	for _, date := range dates {
		entry := database.DailyNoteEntry{
			Date: date,
		}

		// Get note for this day (optional)
		entry.Note, _ = h.DB.GetNoteForDay(ctx, userID, date)

		// Get todos for this day
		entry.Todos, _ = h.DB.GetTodosForDayOnly(ctx, userID, date)

		// Get images for this day
		entry.Images, _ = h.DB.GetImagesForDay(ctx, userID, date)

		entries = append(entries, entry)
	}

	return Render(c, http.StatusOK, pages.DailyNotesPage(user, entries, year, month))
}

func parseIntParam(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
