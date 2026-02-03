package handlers

import (
	"net/http"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminDashboard renders the admin dashboard page
func (h *Handler) AdminDashboard(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user from context (set by RequireAdmin middleware)
	user, ok := c.Get("user").(*database.User)
	if !ok {
		userID, _ := middleware.GetUserID(c)
		var err error
		user, err = h.DB.GetUserByID(ctx, userID)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
	}

	// Get admin stats
	stats, err := h.DB.GetAdminStats(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "خطأ في جلب الإحصائيات")
	}

	// Get all users stats
	usersStats, err := h.DB.GetAllUsersStats(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "خطأ في جلب بيانات المستخدمين")
	}

	return Render(c, http.StatusOK, pages.AdminDashboard(user, stats, usersStats))
}

// DeleteUser deletes a user and all their data
func (h *Handler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()

	// Get current admin user
	adminID, _ := middleware.GetUserID(c)

	// Parse target user ID
	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف المستخدم غير صالح"})
	}

	// Prevent self-deletion
	if targetID == adminID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "لا يمكنك حذف نفسك"})
	}

	// Check if target user exists and is not an admin
	targetUser, err := h.DB.GetUserByID(ctx, targetID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "المستخدم غير موجود"})
	}

	// Prevent deleting other admins
	if targetUser.Role == 1 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "لا يمكن حذف مسؤول آخر"})
	}

	// Delete user and all their data
	err = h.DB.DeleteUser(ctx, targetID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "فشل حذف المستخدم"})
	}

	// Return success with redirect header for HTMX
	c.Response().Header().Set("HX-Redirect", "/admin")
	return c.JSON(http.StatusOK, map[string]string{"message": "تم حذف المستخدم بنجاح"})
}
