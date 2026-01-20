package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ohabits/internal/database"
	"ohabits/internal/middleware"
	"ohabits/templates/pages"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BlogPage renders the blog listing page
func (h *Handler) BlogPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Check for search query
	query := c.QueryParam("q")
	var posts []database.MarkdownNote
	if query != "" {
		posts, err = h.DB.SearchBlogPosts(c.Request().Context(), userID, query)
	} else {
		posts, err = h.DB.GetBlogPosts(c.Request().Context(), userID)
	}
	if err != nil {
		posts = nil
	}

	return Render(c, http.StatusOK, pages.BlogListPage(user, posts, query))
}

// BlogNewPage renders the new blog post page
func (h *Handler) BlogNewPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Create a new empty post
	post, err := h.DB.CreateBlogPost(c.Request().Context(), userID, "مدونة جديدة")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Redirect to edit page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/blog/%s/edit", post.ID.String()))
}

// BlogEditPage renders the blog edit page
func (h *Handler) BlogEditPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/blog")
	}

	post, err := h.DB.GetBlogPost(c.Request().Context(), userID, postID)
	if err != nil || post == nil {
		return c.Redirect(http.StatusSeeOther, "/blog")
	}

	return Render(c, http.StatusOK, pages.BlogEditPage(user, post))
}

// BlogViewPage renders a single blog post
func (h *Handler) BlogViewPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/blog")
	}

	post, err := h.DB.GetBlogPost(c.Request().Context(), userID, postID)
	if err != nil || post == nil {
		return c.Redirect(http.StatusSeeOther, "/blog")
	}

	return Render(c, http.StatusOK, pages.BlogViewPage(user, post))
}

// BlogSave saves a blog post (for auto-save and manual save)
func (h *Handler) BlogSave(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	title := c.FormValue("title")
	content := c.FormValue("content")

	if title == "" {
		title = "مدونة بدون عنوان"
	}

	post, err := h.DB.UpdateBlogPost(c.Request().Context(), userID, postID, title, content)
	if err != nil || post == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"updated_at": post.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// BlogDelete deletes a blog post
func (h *Handler) BlogDelete(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	err = h.DB.DeleteBlogPost(c.Request().Context(), userID, postID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	c.Response().Header().Set("HX-Redirect", "/blog")
	return c.JSON(http.StatusOK, map[string]string{"success": "true"})
}

// BlogSearch handles HTMX search requests for blog posts
func (h *Handler) BlogSearch(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.String(http.StatusUnauthorized, "غير مصرح")
	}

	query := c.QueryParam("q")
	var posts []database.MarkdownNote
	var err error

	if query != "" {
		posts, err = h.DB.SearchBlogPosts(c.Request().Context(), userID, query)
	} else {
		posts, err = h.DB.GetBlogPosts(c.Request().Context(), userID)
	}
	if err != nil {
		posts = nil
	}

	return Render(c, http.StatusOK, pages.BlogPostsList(posts, query))
}

// BlogUploadImage handles image upload for blog posts
func (h *Handler) BlogUploadImage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "لم يتم تحديد صورة"})
	}

	// Validate file size (max 5MB)
	const maxSize = 5 << 20 // 5MB
	if file.Size > maxSize {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "حجم الصورة كبير جداً (الحد الأقصى 5MB)"})
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "نوع الملف غير مدعوم"})
	}

	// Create upload directory
	uploadDir := fmt.Sprintf("uploads/blog/%s", userID.String())
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Return the URL
	imageURL := "/" + filePath
	return c.JSON(http.StatusOK, map[string]string{
		"url": imageURL,
	})
}
