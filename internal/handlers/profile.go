package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Image format decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/image/webp"
)

const (
	avatarMaxSize = 5 << 20 // 5MB
	avatarSize    = 200     // Avatar will be resized to 200x200
)

// ProfilePage renders the profile page
func (h *Handler) ProfilePage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	return Render(c, http.StatusOK, pages.ProfilePage(user, "", ""))
}

// UpdateProfileInfo updates user display name and email
func (h *Handler) UpdateProfileInfo(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	displayName := strings.TrimSpace(c.FormValue("display_name"))
	email := strings.TrimSpace(c.FormValue("email"))

	// Validate input
	if displayName == "" {
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "الاسم مطلوب"))
	}
	if email == "" {
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "البريد الإلكتروني مطلوب"))
	}

	// Check if email is already used by another user
	if email != user.Email {
		exists, err := h.DB.CheckEmailExists(c.Request().Context(), email, userID)
		if err != nil {
			log.Printf("Error checking email: %v", err)
			return Render(c, http.StatusOK, pages.ProfilePage(user, "", "حدث خطأ"))
		}
		if exists {
			return Render(c, http.StatusOK, pages.ProfilePage(user, "", "البريد الإلكتروني مستخدم من قبل"))
		}
	}

	// Update user info
	if err := h.DB.UpdateUserInfo(c.Request().Context(), userID, displayName, email); err != nil {
		log.Printf("Error updating user info: %v", err)
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "حدث خطأ في الحفظ"))
	}

	// Refresh user data
	user, _ = h.DB.GetUserByID(c.Request().Context(), userID)

	return Render(c, http.StatusOK, pages.ProfilePage(user, "تم حفظ التغييرات بنجاح", ""))
}

// UpdateProfilePassword updates user password
func (h *Handler) UpdateProfilePassword(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	// Validate input
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "جميع الحقول مطلوبة"))
	}

	if len(newPassword) < 6 {
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "كلمة المرور يجب أن تكون 6 أحرف على الأقل"))
	}

	if newPassword != confirmPassword {
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "كلمة المرور الجديدة غير متطابقة"))
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "كلمة المرور الحالية غير صحيحة"))
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "حدث خطأ"))
	}

	// Update password
	if err := h.DB.UpdateUserPassword(c.Request().Context(), userID, string(hashedPassword)); err != nil {
		log.Printf("Error updating password: %v", err)
		return Render(c, http.StatusOK, pages.ProfilePage(user, "", "حدث خطأ في الحفظ"))
	}

	return Render(c, http.StatusOK, pages.ProfilePage(user, "تم تغيير كلمة المرور بنجاح", ""))
}

// UpdateProfileAvatar handles avatar upload
func (h *Handler) UpdateProfileAvatar(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "خطأ"})
	}

	// Get uploaded file
	file, err := c.FormFile("avatar")
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Check file size
	if file.Size > avatarMaxSize {
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Check mime type
	mimeType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_format_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Create avatars directory
	avatarsDir := filepath.Join(uploadsDir, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		log.Printf("Error creating avatars directory: %v", err)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"profile_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Open source file
	src, err := file.Open()
	if err != nil {
		log.Printf("Error opening uploaded file: %v", err)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}
	defer src.Close()

	// Decode image directly from the uploaded file
	img, err := imaging.Decode(src)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_format_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Resize to square avatar
	avatar := imaging.Fill(img, avatarSize, avatarSize, imaging.Center, imaging.Lanczos)

	// Generate unique filename (always save as jpg for consistency)
	newFilename := fmt.Sprintf("%s_%s.jpg", userID.String(), uuid.New().String()[:8])
	avatarPath := filepath.Join(avatarsDir, newFilename)

	// Save processed avatar as JPEG
	if err := imaging.Save(avatar, avatarPath, imaging.JPEGQuality(85)); err != nil {
		log.Printf("Error saving avatar: %v", err)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Delete old avatar if exists (with path validation)
	if user.GetAvatarURL() != "" {
		oldPath := strings.TrimPrefix(user.GetAvatarURL(), "/")
		// تنظيف المسار ومنع Path Traversal
		cleanPath := filepath.Clean(oldPath)
		// التأكد من أن المسار داخل مجلد uploads فقط
		if strings.HasPrefix(cleanPath, "uploads/") && !strings.Contains(cleanPath, "..") {
			os.Remove(cleanPath)
		}
	}

	// Update database
	avatarURL := "/" + avatarPath
	if err := h.DB.UpdateUserAvatar(c.Request().Context(), userID, avatarURL); err != nil {
		os.Remove(avatarPath)
		log.Printf("Error updating avatar in DB: %v", err)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"save_error","type":"error"}}`)
		return Render(c, http.StatusOK, pages.AvatarSection(user))
	}

	// Refresh user data
	user, _ = h.DB.GetUserByID(c.Request().Context(), userID)

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"avatar_saved","type":"success"}}`)
	return Render(c, http.StatusOK, pages.AvatarSection(user))
}

// MARK: - Profile API (للتطبيق الأصلي iOS/macOS)

// ProfileAPIResponse represents the JSON response for profile API
type ProfileAPIResponse struct {
	Status string       `json:"status"`
	User   *ProfileUser `json:"user,omitempty"`
	Error  string       `json:"error,omitempty"`
}

// ProfileUser represents user data in API response
type ProfileUser struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	FullName        string  `json:"full_name"`
	ProfileImageUrl *string `json:"profile_image_url"`
	Role            int     `json:"role"`
}

// ProfileImageAPIResponse represents the response for profile image upload
type ProfileImageAPIResponse struct {
	Status   string  `json:"status"`
	ImageUrl *string `json:"image_url,omitempty"`
	Error    string  `json:"error,omitempty"`
}

// GetProfileAPI returns user profile as JSON
// GET /api/user/profile
func (h *Handler) GetProfileAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ProfileAPIResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ProfileAPIResponse{
			Status: "error",
			Error:  "User not found",
		})
	}

	return c.JSON(http.StatusOK, ProfileAPIResponse{
		Status: "success",
		User: &ProfileUser{
			ID:              user.ID.String(),
			Email:           user.Email,
			FullName:        user.DisplayName,
			ProfileImageUrl: user.AvatarURL,
			Role:            user.Role,
		},
	})
}

// UpdateProfileRequest represents the request body for profile update
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
}

// UpdateProfileAPI updates user profile (name only)
// PUT /api/user/profile
func (h *Handler) UpdateProfileAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ProfileAPIResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ProfileAPIResponse{
			Status: "error",
			Error:  "Invalid request format",
		})
	}

	displayName := strings.TrimSpace(req.FullName)
	if displayName == "" {
		return c.JSON(http.StatusBadRequest, ProfileAPIResponse{
			Status: "error",
			Error:  "Name is required",
		})
	}

	// Get current user to preserve email
	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ProfileAPIResponse{
			Status: "error",
			Error:  "User not found",
		})
	}

	// Update only the display name
	if err := h.DB.UpdateUserInfo(c.Request().Context(), userID, displayName, user.Email); err != nil {
		log.Printf("Error updating user profile: %v", err)
		return c.JSON(http.StatusInternalServerError, ProfileAPIResponse{
			Status: "error",
			Error:  "Failed to update profile",
		})
	}

	// Refresh user data
	user, _ = h.DB.GetUserByID(c.Request().Context(), userID)

	return c.JSON(http.StatusOK, ProfileAPIResponse{
		Status: "success",
		User: &ProfileUser{
			ID:              user.ID.String(),
			Email:           user.Email,
			FullName:        user.DisplayName,
			ProfileImageUrl: user.AvatarURL,
			Role:            user.Role,
		},
	})
}

// UploadProfileImageAPI handles profile image upload from native app
// POST /api/user/profile/image
func (h *Handler) UploadProfileImageAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "User not found",
		})
	}

	// Get uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ProfileImageAPIResponse{
			Status: "error",
			Error:  "No image provided",
		})
	}

	// Check file size
	if file.Size > avatarMaxSize {
		return c.JSON(http.StatusBadRequest, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Image too large (max 5MB)",
		})
	}

	// Check mime type
	mimeType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		return c.JSON(http.StatusBadRequest, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Invalid image format",
		})
	}

	// Create avatars directory
	avatarsDir := filepath.Join(uploadsDir, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		log.Printf("Error creating avatars directory: %v", err)
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Server error",
		})
	}

	// Open source file
	src, err := file.Open()
	if err != nil {
		log.Printf("Error opening uploaded file: %v", err)
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Failed to process image",
		})
	}
	defer src.Close()

	// Decode image
	img, err := imaging.Decode(src)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		return c.JSON(http.StatusBadRequest, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Invalid image format",
		})
	}

	// Resize to square avatar
	avatar := imaging.Fill(img, avatarSize, avatarSize, imaging.Center, imaging.Lanczos)

	// Generate unique filename
	newFilename := fmt.Sprintf("%s_%s.jpg", userID.String(), uuid.New().String()[:8])
	avatarPath := filepath.Join(avatarsDir, newFilename)

	// Save processed avatar as JPEG
	if err := imaging.Save(avatar, avatarPath, imaging.JPEGQuality(85)); err != nil {
		log.Printf("Error saving avatar: %v", err)
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Failed to save image",
		})
	}

	// Delete old avatar if exists
	if user.GetAvatarURL() != "" {
		oldPath := strings.TrimPrefix(user.GetAvatarURL(), "/")
		cleanPath := filepath.Clean(oldPath)
		if strings.HasPrefix(cleanPath, "uploads/") && !strings.Contains(cleanPath, "..") {
			os.Remove(cleanPath)
		}
	}

	// Update database
	avatarURL := "/" + avatarPath
	if err := h.DB.UpdateUserAvatar(c.Request().Context(), userID, avatarURL); err != nil {
		os.Remove(avatarPath)
		log.Printf("Error updating avatar in DB: %v", err)
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Failed to save image",
		})
	}

	return c.JSON(http.StatusOK, ProfileImageAPIResponse{
		Status:   "success",
		ImageUrl: &avatarURL,
	})
}

// DeleteProfileImageAPI deletes the user's profile image
// DELETE /api/user/profile/image
func (h *Handler) DeleteProfileImageAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "User not found",
		})
	}

	// Delete old avatar if exists
	if user.GetAvatarURL() != "" {
		oldPath := strings.TrimPrefix(user.GetAvatarURL(), "/")
		cleanPath := filepath.Clean(oldPath)
		if strings.HasPrefix(cleanPath, "uploads/") && !strings.Contains(cleanPath, "..") {
			os.Remove(cleanPath)
		}
	}

	// Update database to clear avatar
	if err := h.DB.UpdateUserAvatar(c.Request().Context(), userID, ""); err != nil {
		log.Printf("Error clearing avatar in DB: %v", err)
		return c.JSON(http.StatusInternalServerError, ProfileImageAPIResponse{
			Status: "error",
			Error:  "Failed to delete image",
		})
	}

	return c.JSON(http.StatusOK, ProfileImageAPIResponse{
		Status: "success",
	})
}
