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
	// Check if already logged in
	if cookie, err := c.Cookie("token"); err == nil {
		if _, err := h.Auth.ValidateToken(cookie.Value); err == nil {
			return c.Redirect(http.StatusSeeOther, "/")
		}
	}

	return Render(c, http.StatusOK, pages.Signup("", ""))
}

// Signup handles signup form submission
func (h *Handler) Signup(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")
	displayName := c.FormValue("display_name")

	if email == "" || password == "" {
		return Render(c, http.StatusOK, pages.Signup("البريد الإلكتروني وكلمة المرور مطلوبان", email))
	}

	if password != confirmPassword {
		return Render(c, http.StatusOK, pages.Signup("كلمتا المرور غير متطابقتان", email))
	}

	if len(password) < 6 {
		return Render(c, http.StatusOK, pages.Signup("كلمة المرور يجب أن تكون ٦ أحرف على الأقل", email))
	}

	if displayName == "" {
		displayName = email
	}

	user, err := h.DB.CreateUser(c.Request().Context(), email, password, displayName)
	if err != nil {
		if err == database.ErrEmailExists {
			return Render(c, http.StatusOK, pages.Signup("البريد الإلكتروني مستخدم بالفعل", email))
		}
		return Render(c, http.StatusOK, pages.Signup("حدث خطأ، حاول مرة أخرى", email))
	}

	// Generate JWT token
	token, err := h.Auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return Render(c, http.StatusOK, pages.Signup("حدث خطأ في إنشاء الحساب", email))
	}

	// Set cookie
	middleware.SetAuthCookie(c, token)

	return c.Redirect(http.StatusSeeOther, "/")
}

// Logout handles logout
func (h *Handler) Logout(c echo.Context) error {
	middleware.ClearAuthCookie(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}
