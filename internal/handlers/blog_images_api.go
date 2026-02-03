package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ohabits/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UploadBlogImageAPI handles blog image upload from native app
func (h *Handler) UploadBlogImageAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "غير مصرح"})
	}

	// Get note_id (server ID of the markdown note)
	noteIDStr := c.FormValue("note_id")
	if noteIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "note_id مطلوب"})
	}
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "note_id غير صالح"})
	}

	// Get position marker
	positionMarker := c.FormValue("position_marker")
	if positionMarker == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "position_marker مطلوب"})
	}

	// Verify the note belongs to the user
	post, err := h.DB.GetBlogPost(c.Request().Context(), userID, noteID)
	if err != nil || post == nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"status": "error", "error": "المدونة غير موجودة"})
	}

	// Get image file
	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "لم يتم تحديد صورة"})
	}

	// Validate file size (max 10MB)
	const maxSize = 10 << 20
	if file.Size > maxSize {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "حجم الصورة كبير جداً"})
	}

	// Validate file type
	mimeType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "نوع الملف غير مدعوم"})
	}

	// Create upload directories
	userDir := filepath.Join("uploads", "blog", userID.String())
	originalsDir := filepath.Join(userDir, "originals")
	thumbsDir := filepath.Join(userDir, "thumbnails")

	if err := os.MkdirAll(originalsDir, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في إنشاء المجلدات"})
	}
	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في إنشاء المجلدات"})
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	newFilename := fmt.Sprintf("%s_%s%s", positionMarker, uuid.New().String()[:8], ext)

	originalPath := filepath.Join(originalsDir, newFilename)
	thumbnailPath := filepath.Join(thumbsDir, newFilename)

	// Save original file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في قراءة الملف"})
	}
	defer src.Close()

	dst, err := os.Create(originalPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في حفظ الملف"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		os.Remove(originalPath)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في حفظ الملف"})
	}

	// Create thumbnail (optional)
	createThumbnail(originalPath, thumbnailPath)

	// Save to database
	img, err := h.DB.SaveBlogImage(
		c.Request().Context(),
		userID,
		noteID,
		"/"+originalPath,
		"/"+thumbnailPath,
		file.Filename,
		mimeType,
		int(file.Size),
		positionMarker,
	)
	if err != nil {
		os.Remove(originalPath)
		os.Remove(thumbnailPath)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في حفظ السجل"})
	}

	thumbPathStr := ""
	if img.ThumbnailPath != nil {
		thumbPathStr = *img.ThumbnailPath
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"image": map[string]interface{}{
			"id":              img.ID,
			"original_path":   img.OriginalPath,
			"thumbnail_path":  thumbPathStr,
			"filename":        img.Filename,
			"mime_type":       img.MimeType,
			"size_bytes":      img.SizeBytes,
			"position_marker": img.PositionMarker,
		},
	})
}

// DeleteBlogImageAPI handles blog image deletion from native app
func (h *Handler) DeleteBlogImageAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "غير مصرح"})
	}

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "معرف غير صالح"})
	}

	// Soft delete from database
	img, err := h.DB.DeleteBlogImage(c.Request().Context(), imageID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "خطأ في الحذف"})
	}

	// Delete files (optional)
	os.Remove(strings.TrimPrefix(img.OriginalPath, "/"))
	if img.ThumbnailPath != nil {
		os.Remove(strings.TrimPrefix(*img.ThumbnailPath, "/"))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}
