package handlers

import (
	"net/http"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/labstack/echo/v4"
)

// LoginPage renders the login page
func (h *Handler) LoginPage(c echo.Context) error {
	// Check if already logged in
	if cookie, err := c.Cookie("token"); err == nil {
		if _, err := h.Auth.ValidateToken(cookie.Value); err == nil {
			return c.Redirect(http.StatusSeeOther, "/")
		}
	}

	return Render(c, http.StatusOK, pages.Login("", ""))
}

// Login handles login form submission
func (h *Handler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return Render(c, http.StatusOK, pages.Login("البريد الإلكتروني وكلمة المرور مطلوبان", email))
	}

	user, err := h.DB.AuthenticateUser(c.Request().Context(), email, password)
	if err != nil {
		if err == database.ErrUserNotFound || err == database.ErrInvalidPassword {
			return Render(c, http.StatusOK, pages.Login("البريد الإلكتروني أو كلمة المرور غير صحيحة", email))
		}
		return Render(c, http.StatusOK, pages.Login("حدث خطأ، حاول مرة أخرى", email))
	}

	// Generate JWT token
	token, err := h.Auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return Render(c, http.StatusOK, pages.Login("حدث خطأ في تسجيل الدخول", email))
	}

	// Set cookie
	middleware.SetAuthCookie(c, token)

	return c.Redirect(http.StatusSeeOther, "/")
}

// SignupPage renders the signup page
func (h *Handler) SignupPage(c echo.Context) error {
	return Render(c, http.StatusOK, pages.Login("التسجيل مغلق حالياً", ""))
}

// Signup handles signup form submission
func (h *Handler) Signup(c echo.Context) error {
	return Render(c, http.StatusOK, pages.Login("التسجيل مغلق حالياً", ""))
}

// Logout handles logout
func (h *Handler) Logout(c echo.Context) error {
	middleware.ClearAuthCookie(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// APILogin handles API-based email/password login for native apps
func (h *Handler) APILogin(c echo.Context) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}

	user, err := h.DB.AuthenticateUser(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if err == database.ErrUserNotFound || err == database.ErrInvalidPassword {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication failed",
		})
	}

	// Generate JWT token
	token, err := h.Auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"fullName": user.DisplayName,
		},
		"isNewUser": false,
	})
}
