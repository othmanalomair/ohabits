package handlers

import (
	"net/http"
	"strings"
	"time"

	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CalendarPage renders the calendar management page
func (h *Handler) CalendarPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	events, _ := h.DB.GetCalendarEventsByUserID(c.Request().Context(), userID)

	return Render(c, http.StatusOK, pages.CalendarPage(user, events))
}

// CreateCalendarEvent creates a new calendar event
func (h *Handler) CreateCalendarEvent(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "العنوان مطلوب"})
	}

	eventType := c.FormValue("event_type")
	if eventType == "" {
		eventType = "anniversary"
	}

	// Validate event type
	validTypes := map[string]bool{
		"birthday":    true,
		"travel":      true,
		"holiday":     true,
		"anniversary": true,
		"general":     true,
	}
	if !validTypes[eventType] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "نوع الحدث غير صالح"})
	}

	dateStr := c.FormValue("event_date")
	eventDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "التاريخ غير صالح"})
	}

	// Parse end_date (optional)
	var endDate *time.Time
	endDateStr := c.FormValue("end_date")
	if endDateStr != "" {
		ed, err := time.Parse("2006-01-02", endDateStr)
		if err == nil && !ed.Before(eventDate) {
			endDate = &ed
		}
	}

	isRecurring := c.FormValue("is_recurring") == "on" || c.FormValue("is_recurring") == "true"
	notes := strings.TrimSpace(c.FormValue("notes"))

	_, err = h.DB.CreateCalendarEvent(c.Request().Context(), userID, title, eventType, eventDate, endDate, isRecurring, notes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"event_saved","type":"success"}}`)

	// Return updated events list
	events, _ := h.DB.GetCalendarEventsByUserID(c.Request().Context(), userID)
	return Render(c, http.StatusOK, pages.CalendarEventsList(events))
}

// UpdateCalendarEvent updates an existing calendar event
func (h *Handler) UpdateCalendarEvent(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "العنوان مطلوب"})
	}

	eventType := c.FormValue("event_type")
	if eventType == "" {
		eventType = "anniversary"
	}

	dateStr := c.FormValue("event_date")
	eventDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "التاريخ غير صالح"})
	}

	// Parse end_date (optional)
	var endDate *time.Time
	endDateStr := c.FormValue("end_date")
	if endDateStr != "" {
		ed, err := time.Parse("2006-01-02", endDateStr)
		if err == nil && !ed.Before(eventDate) {
			endDate = &ed
		}
	}

	isRecurring := c.FormValue("is_recurring") == "on" || c.FormValue("is_recurring") == "true"
	notes := strings.TrimSpace(c.FormValue("notes"))

	err = h.DB.UpdateCalendarEvent(c.Request().Context(), eventID, title, eventType, eventDate, endDate, isRecurring, notes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"event_saved","type":"success"}}`)

	// Return updated events list
	events, _ := h.DB.GetCalendarEventsByUserID(c.Request().Context(), userID)
	return Render(c, http.StatusOK, pages.CalendarEventsList(events))
}

// DeleteCalendarEvent deletes a calendar event
func (h *Handler) DeleteCalendarEvent(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	if err := h.DB.DeleteCalendarEvent(c.Request().Context(), eventID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"event_deleted","type":"success"}}`)

	// Return updated events list
	events, _ := h.DB.GetCalendarEventsByUserID(c.Request().Context(), userID)
	return Render(c, http.StatusOK, pages.CalendarEventsList(events))
}
