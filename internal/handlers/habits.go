package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ohabits/internal/middleware"
	"ohabits/templates/pages"
	"ohabits/templates/partials"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// HabitsPage renders the habits management page
func (h *Handler) HabitsPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	habits, _ := h.DB.GetHabitsByUserID(c.Request().Context(), userID)

	return Render(c, http.StatusOK, pages.HabitsPage(user, habits))
}

// ToggleHabit toggles habit completion
func (h *Handler) ToggleHabit(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	habitIDStr := c.Param("id")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	completed, err := h.DB.ToggleHabitCompletion(c.Request().Context(), userID, habitID, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Get updated habits list
	habits, _ := h.DB.GetHabitsForDay(c.Request().Context(), userID, date)

	// Find the toggled habit
	for _, habit := range habits {
		if habit.ID == habitID {
			return Render(c, http.StatusOK, partials.HabitItem(habit, date, completed))
		}
	}

	return c.NoContent(http.StatusOK)
}

// dayNames maps form index to day name (form uses 0=Sunday order)
var dayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

// CreateHabit creates a new habit
func (h *Handler) CreateHabit(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	name := c.FormValue("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	// Parse scheduled days from form and convert to day names
	var scheduledDays []string
	for i := 0; i < 7; i++ {
		if c.FormValue("day_"+strconv.Itoa(i)) == "on" {
			scheduledDays = append(scheduledDays, dayNames[i])
		}
	}

	_, err := h.DB.CreateHabit(c.Request().Context(), userID, name, scheduledDays)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"habit_saved","type":"success"}}`)

	// Check if request is from habits management page
	referer := c.Request().Referer()
	if len(referer) >= 7 && contains(referer, "/habits") {
		habits, _ := h.DB.GetHabitsByUserID(c.Request().Context(), userID)
		return Render(c, http.StatusOK, pages.HabitsManageList(habits))
	}

	// Return updated habits list for dashboard
	date := time.Now()
	habits, _ := h.DB.GetHabitsForDay(c.Request().Context(), userID, date)

	return Render(c, http.StatusOK, partials.HabitsList(habits, date))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DeleteHabit deletes a habit
func (h *Handler) DeleteHabit(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	habitIDStr := c.Param("id")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	if err := h.DB.DeleteHabit(c.Request().Context(), habitID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"habit_deleted","type":"success"}}`)

	// Check if request is from habits management page
	currentURL := c.Request().Header.Get("HX-Current-URL")
	referer := c.Request().Referer()
	if contains(currentURL, "/habits") || contains(referer, "/habits") {
		habits, _ := h.DB.GetHabitsByUserID(c.Request().Context(), userID)
		return Render(c, http.StatusOK, pages.HabitsManageList(habits))
	}

	// Return updated habits list for dashboard
	date := time.Now()
	habits, _ := h.DB.GetHabitsForDay(c.Request().Context(), userID, date)

	return Render(c, http.StatusOK, partials.HabitsList(habits, date))
}

// UpdateHabit updates a habit
func (h *Handler) UpdateHabit(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	habitIDStr := c.Param("id")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	name := c.FormValue("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	// Parse scheduled days from form
	var scheduledDays []string
	for i := 0; i < 7; i++ {
		if c.FormValue("day_"+strconv.Itoa(i)) == "on" {
			scheduledDays = append(scheduledDays, dayNames[i])
		}
	}

	if err := h.DB.UpdateHabit(c.Request().Context(), habitID, name, scheduledDays); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"habit_saved","type":"success"}}`)

	// Get updated habit and return
	habits, _ := h.DB.GetHabitsByUserID(c.Request().Context(), userID)
	for _, habit := range habits {
		if habit.ID == habitID {
			return Render(c, http.StatusOK, pages.HabitManageItem(habit))
		}
	}

	return c.NoContent(http.StatusOK)
}
