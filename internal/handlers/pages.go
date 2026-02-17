package handlers

import (
	"net/http"

	"ohabits/templates/pages"

	"github.com/labstack/echo/v4"
)

func (h *Handler) PrivacyPage(c echo.Context) error {
	return Render(c, http.StatusOK, pages.Privacy())
}

func (h *Handler) TermsPage(c echo.Context) error {
	return Render(c, http.StatusOK, pages.Terms())
}

func (h *Handler) SupportPage(c echo.Context) error {
	return Render(c, http.StatusOK, pages.Support())
}
