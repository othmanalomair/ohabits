package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"ohabits/internal/middleware"
	"ohabits/templates/partials"

	"github.com/labstack/echo/v4"
)

// SaveNote saves a note for a day
func (h *Handler) SaveNote(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	text := c.FormValue("text")
	dateStr := c.FormValue("date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	note, err := h.DB.SaveNote(c.Request().Context(), userID, text, date)
	if err != nil {
		// Send error toast
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"save_error","type":"error"}}`)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Get images for this day
	images, _ := h.DB.GetImagesForDay(c.Request().Context(), userID, date)

	// Send success toast - use code that JS will translate
	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"note_saved","type":"success"}}`)

	return Render(c, http.StatusOK, partials.NoteSection(note, images, date))
}

// SaveMood saves mood rating for a day
func (h *Handler) SaveMood(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	ratingStr := c.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "تقييم غير صالح"})
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	mood, err := h.DB.SaveMood(c.Request().Context(), userID, rating, date)
	if err != nil {
		log.Printf("SaveMood error: %v", err)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"save_error","type":"error"}}`)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Send success toast
	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"mood_saved","type":"success"}}`)

	return Render(c, http.StatusOK, partials.MoodSection(mood, date))
}
