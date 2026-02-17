package handlers

import (
	"net/http"

	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/labstack/echo/v4"
)

// LandingOrDashboard shows the landing page for unauthenticated users,
// or the dashboard for authenticated users.
func (h *Handler) LandingOrDashboard(c echo.Context) error {
	_, ok := middleware.GetUserID(c)
	if ok {
		return h.Dashboard(c)
	}
	return Render(c, http.StatusOK, pages.Landing())
}
