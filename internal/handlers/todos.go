package handlers

import (
	"net/http"
	"time"

	"ohabits/internal/middleware"
	"ohabits/templates/partials"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CreateTodo creates a new todo
func (h *Handler) CreateTodo(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	text := c.FormValue("text")
	if text == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "النص مطلوب"})
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	_, err = h.DB.CreateTodo(c.Request().Context(), userID, text, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Return updated list
	todos, _ := h.DB.GetTodosForDay(c.Request().Context(), userID, date)
	return Render(c, http.StatusOK, partials.TodosList(todos, date))
}

// ToggleTodo toggles todo completion
func (h *Handler) ToggleTodo(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	todoIDStr := c.Param("id")
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	completed, err := h.DB.ToggleTodo(c.Request().Context(), todoID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	dateStr := c.FormValue("date")
	date, _ := time.Parse("2006-01-02", dateStr)
	if date.IsZero() {
		date = time.Now()
	}

	// Get updated todos
	todos, _ := h.DB.GetTodosForDay(c.Request().Context(), userID, date)

	// Find the toggled todo
	for _, todo := range todos {
		if todo.ID == todoID {
			return Render(c, http.StatusOK, partials.TodoItem(todo, date, completed))
		}
	}

	return c.NoContent(http.StatusOK)
}

// DeleteTodo deletes a todo
func (h *Handler) DeleteTodo(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	todoIDStr := c.Param("id")
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	if err := h.DB.DeleteTodo(c.Request().Context(), todoID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Try query param first, then form value
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		dateStr = c.FormValue("date")
	}
	date, _ := time.Parse("2006-01-02", dateStr)
	if date.IsZero() {
		date = time.Now()
	}

	// Return updated list
	todos, _ := h.DB.GetTodosForDay(c.Request().Context(), userID, date)
	return Render(c, http.StatusOK, partials.TodosList(todos, date))
}
