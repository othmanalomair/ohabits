package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/pages"
	"ohabits/templates/partials"

	"github.com/google/uuid"
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

	// Parse cardio as array
	var cardio []database.Cardio
	if cardioName != "" {
		minutes, _ := strconv.Atoi(cardioMinutesStr)
		cardio = []database.Cardio{{
			Name:    cardioName,
			Minutes: minutes,
		}}
	} else {
		cardio = []database.Cardio{}
	}

	// Get exercises from the selected workout
	exercises := []database.Exercise{}
	if workoutName != "" {
		workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)
		for _, w := range workouts {
			if w.Name == workoutName {
				exercises = w.Exercises
				break
			}
		}
	}

	log, err := h.DB.SaveWorkoutLog(c.Request().Context(), userID, workoutName, exercises, cardio, weight, date)
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"save_error","type":"error"}}`)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ: " + err.Error()})
	}

	// Send success toast
	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"workout_saved","type":"success"}}`)

	// Get workouts for select options
	workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)

	return Render(c, http.StatusOK, partials.WorkoutSection(workouts, log, date))
}

// WorkoutsPage renders the workouts management page
func (h *Handler) WorkoutsPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)

	return Render(c, http.StatusOK, pages.WorkoutsPage(user, workouts))
}

// CreateWorkout creates a new workout
func (h *Handler) CreateWorkout(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	name := c.FormValue("name")
	day := c.FormValue("day")

	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	if day == "" {
		day = "N/A"
	}

	// Parse exercises from form
	var exercises []database.Exercise
	for i := 1; i <= 50; i++ {
		exName := c.FormValue("exercise_" + strconv.Itoa(i))
		if exName != "" {
			exercises = append(exercises, database.Exercise{
				Order: i,
				Name:  exName,
			})
		}
	}

	_, err := h.DB.CreateWorkout(c.Request().Context(), userID, name, day, exercises)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ: " + err.Error()})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"workout_saved","type":"success"}}`)

	// Return updated list
	workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)
	return Render(c, http.StatusOK, pages.WorkoutsManageList(workouts))
}

// UpdateWorkout updates a workout
func (h *Handler) UpdateWorkout(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	workoutIDStr := c.Param("id")
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	name := c.FormValue("name")
	day := c.FormValue("day")

	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	if day == "" {
		day = "N/A"
	}

	// Parse exercises from form
	var exercises []database.Exercise
	for i := 1; i <= 50; i++ {
		exName := c.FormValue("exercise_" + strconv.Itoa(i))
		if exName != "" {
			exercises = append(exercises, database.Exercise{
				Order: i,
				Name:  strings.TrimSpace(exName),
			})
		}
	}

	if err := h.DB.UpdateWorkout(c.Request().Context(), workoutID, name, day, exercises); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ: " + err.Error()})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"workout_saved","type":"success"}}`)

	// Get updated workout and return
	workout, err := h.DB.GetWorkoutByID(c.Request().Context(), workoutID)
	if err != nil {
		workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)
		return Render(c, http.StatusOK, pages.WorkoutsManageList(workouts))
	}

	return Render(c, http.StatusOK, pages.WorkoutManageItem(*workout))
}

// DeleteWorkout deletes a workout
func (h *Handler) DeleteWorkout(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	workoutIDStr := c.Param("id")
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	if err := h.DB.DeleteWorkout(c.Request().Context(), workoutID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"workout_deleted","type":"success"}}`)

	// Return updated list
	workouts, _ := h.DB.GetWorkoutsByUserID(c.Request().Context(), userID)
	return Render(c, http.StatusOK, pages.WorkoutsManageList(workouts))
}

// ReorderWorkouts reorders workouts based on drag and drop
func (h *Handler) ReorderWorkouts(c echo.Context) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	var req struct {
		IDs []string `json:"ids"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "بيانات غير صالحة"})
	}

	// Convert string IDs to UUIDs
	workoutIDs := make([]uuid.UUID, len(req.IDs))
	for i, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
		}
		workoutIDs[i] = id
	}

	if err := h.DB.ReorderWorkouts(c.Request().Context(), workoutIDs); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
