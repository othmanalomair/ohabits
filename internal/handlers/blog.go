package handlers

import (
	"net/http"

	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/labstack/echo/v4"
)

// BlogPage renders the blog page (coming soon)
func (h *Handler) BlogPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	return Render(c, http.StatusOK, pages.BlogPage(user))
}
