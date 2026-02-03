package middleware

import (
	"net/http"
	"strings"

	"ohabits/internal/database"

	"github.com/labstack/echo/v4"
)

// Role constants
const (
	RoleAdmin      = 1
	RoleNormal     = 2
	RoleSubscribed = 3
)

// RequireAdmin middleware checks if user has admin role
func RequireAdmin(db *database.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := GetUserID(c)
			if !ok {
				return redirectOrJSON(c, "/login", "غير مصرح")
			}

			user, err := db.GetUserByID(c.Request().Context(), userID)
			if err != nil {
				return redirectOrJSON(c, "/login", "المستخدم غير موجود")
			}

			if user.Role != RoleAdmin {
				return redirectOrJSON(c, "/", "غير مصرح بالوصول")
			}

			// Store user in context for use in handlers
			c.Set("user", user)
			return next(c)
		}
	}
}

// RequireSubscribed middleware checks if user has subscribed or admin role
func RequireSubscribed(db *database.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := GetUserID(c)
			if !ok {
				return redirectOrJSON(c, "/login", "غير مصرح")
			}

			user, err := db.GetUserByID(c.Request().Context(), userID)
			if err != nil {
				return redirectOrJSON(c, "/login", "المستخدم غير موجود")
			}

			if user.Role != RoleAdmin && user.Role != RoleSubscribed {
				return redirectOrJSON(c, "/", "هذه الميزة متاحة للمشتركين فقط")
			}

			c.Set("user", user)
			return next(c)
		}
	}
}

// GetUserRole returns the user's role from context
func GetUserRole(c echo.Context) int {
	user, ok := c.Get("user").(*database.User)
	if !ok {
		return RoleNormal
	}
	return user.Role
}

// IsAdmin checks if the user in context is an admin
func IsAdmin(c echo.Context) bool {
	return GetUserRole(c) == RoleAdmin
}

// IsSubscribed checks if the user in context is subscribed or admin
func IsSubscribed(c echo.Context) bool {
	role := GetUserRole(c)
	return role == RoleAdmin || role == RoleSubscribed
}

// redirectOrJSON redirects for web requests or returns JSON for API requests
func redirectOrJSON(c echo.Context, redirectURL, message string) error {
	if strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	return c.JSON(http.StatusForbidden, map[string]string{"error": message})
}
