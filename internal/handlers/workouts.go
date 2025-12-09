package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/partials"

	"github.com/labstack/echo/v4"
)

// SaveWorkoutLog saves a workout log
func (h *Handler) SaveWorkoutLog(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	workoutName := c.FormValue("workout_name")
	weightStr := c.FormValue("weight")
	cardioName := c.FormValue("cardio_name")
	cardioMinutesStr := c.FormValue("cardio_minutes")
	notes := c.FormValue("notes")

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	// Parse weight
	weight := 0.0
	if weightStr != "" {
		weight, _ = strconv.ParseFloat(weightStr, 64)
	}

	// Parse cardio
	var cardio *database.Cardio
	if cardioName != "" {
		minutes, _ := strconv.Atoi(cardioMinutesStr)
		cardio = &database.Cardio{
			Name:    cardioName,
			Minutes: minutes,
		}
	}

	// Get exercises (for now, just save the workout name)
	exercises := []database.Exercise{}

	log, err := h.DB.SaveWorkoutLog(c.Request().Context(), userID, workoutName, exercises, cardio, weight, date, notes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Get workouts for select options
	workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)

	return Render(c, http.StatusOK, partials.WorkoutSection(workouts, log, date))
}
