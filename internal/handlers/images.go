package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/partials"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	uploadsDir    = "uploads"
	thumbWidth    = 150
	thumbHeight   = 150
	maxUploadSize = 10 << 20 // 10MB
)

// isAPIRequest checks if the request is from the native app (API)
func isAPIRequest(c echo.Context) bool {
	return strings.HasPrefix(c.Path(), "/api/")
}

// UploadImages handles multiple image uploads for a day
func (h *Handler) UploadImages(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	// Parse date - try ISO8601 first (from API), then simple date format (from web)
	dateStr := c.FormValue("date")
	var date time.Time
	var err error
	
	if dateStr != "" {
		// Try ISO8601 format first (2026-01-29T12:00:00Z)
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			// Try simple date format (2006-01-02)
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				date = time.Now()
			}
		}
	} else {
		date = time.Now()
	}
	
	// Normalize to just the date (no time)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Get multipart form
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("MultipartForm error: %v", err)
		if isAPIRequest(c) {
			return c.JSON(http.StatusBadRequest, map[string]string{"status": "error", "error": "خطأ في قراءة الملفات"})
		}
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"save_error","type":"error"}}`)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "خطأ في قراءة الملفات"})
	}

	files := form.File["images"]
	if len(files) == 0 {
		if isAPIRequest(c) {
			return c.JSON(http.StatusBadRequest, map[string]string{"status": "error", "error": "لم يتم اختيار صور"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "لم يتم اختيار صور"})
	}

	// Create upload directories
	userDir := filepath.Join(uploadsDir, userID.String())
	originalsDir := filepath.Join(userDir, "originals")
	thumbsDir := filepath.Join(userDir, "thumbnails")

	if err := os.MkdirAll(originalsDir, 0755); err != nil {
		log.Printf("MkdirAll error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "error": "خطأ في إنشاء المجلدات"})
	}
	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		log.Printf("MkdirAll error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "error": "خطأ في إنشاء المجلدات"})
	}

	var savedImages []database.DailyImage

	for _, file := range files {
		// Check file size
		if file.Size > maxUploadSize {
			continue // Skip files that are too large
		}

		// Check mime type
		mimeType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(mimeType, "image/") {
			continue // Skip non-image files
		}

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		newFilename := fmt.Sprintf("%s_%s%s", date.Format("2006-01-02"), uuid.New().String()[:8], ext)

		originalPath := filepath.Join(originalsDir, newFilename)
		thumbnailPath := filepath.Join(thumbsDir, newFilename)

		// Open source file
		src, err := file.Open()
		if err != nil {
			log.Printf("File open error: %v", err)
			continue
		}

		// Save original
		dst, err := os.Create(originalPath)
		if err != nil {
			src.Close()
			log.Printf("Create file error: %v", err)
			continue
		}

		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			log.Printf("Copy error: %v", err)
			os.Remove(originalPath)
			continue
		}

		// Create thumbnail
		if err := createThumbnail(originalPath, thumbnailPath); err != nil {
			log.Printf("Thumbnail error: %v", err)
			// Continue without thumbnail if it fails
			thumbnailPath = originalPath
		}

		// Save to database
		img, err := h.DB.SaveDailyImage(
			c.Request().Context(),
			userID,
			date,
			"/"+originalPath,
			"/"+thumbnailPath,
			file.Filename,
			mimeType,
			int(file.Size),
		)
		if err != nil {
			log.Printf("DB save error: %v", err)
			continue
		}

		savedImages = append(savedImages, *img)
	}

	if len(savedImages) == 0 {
		if isAPIRequest(c) {
			return c.JSON(http.StatusBadRequest, map[string]string{"status": "error", "error": "لم يتم حفظ أي صورة"})
		}
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"save_error","type":"error"}}`)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "لم يتم حفظ أي صورة"})
	}

	// For API requests, return JSON with the first saved image info
	if isAPIRequest(c) {
		img := savedImages[0]
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"image": map[string]interface{}{
				"id":            img.ID,
				"original_path":  img.OriginalPath,
				"thumbnail_path": img.ThumbnailPath,
				"filename":      img.Filename,
				"mime_type":      img.MimeType,
				"size_bytes":     img.SizeBytes,
			},
		})
	}

	// For web requests, return HTML
	images, _ := h.DB.GetImagesForDay(c.Request().Context(), userID, date)
	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"images_saved","type":"success"}}`)
	return Render(c, http.StatusOK, partials.ImageGallery(images, date))
}

// DeleteImage deletes an image
func (h *Handler) DeleteImage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	// Delete from database and get the paths
	img, err := h.DB.DeleteDailyImage(c.Request().Context(), imageID, userID)
	if err != nil {
		log.Printf("Delete error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "خطأ في الحذف"})
	}

	// Delete files (remove leading slash)
	os.Remove(strings.TrimPrefix(img.OriginalPath, "/"))
	os.Remove(strings.TrimPrefix(img.ThumbnailPath, "/"))

	// For API requests, return JSON
	if isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	}

	// For web requests, return HTML
	images, _ := h.DB.GetImagesForDay(c.Request().Context(), userID, date)
	return Render(c, http.StatusOK, partials.ImageGallery(images, date))
}

// createThumbnail creates a thumbnail from the original image
func createThumbnail(srcPath, dstPath string) error {
	src, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// Resize to fit within thumbnail dimensions while maintaining aspect ratio
	thumb := imaging.Fill(src, thumbWidth, thumbHeight, imaging.Center, imaging.Lanczos)

	return imaging.Save(thumb, dstPath)
}
