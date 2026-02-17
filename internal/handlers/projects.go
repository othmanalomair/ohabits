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

// ========== PROJECT HANDLERS ==========

// GetProjects returns all projects for the authenticated user
// GET /api/projects
func (h *Handler) GetProjects(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	ctx := c.Request().Context()
	projects, err := h.DB.GetActiveProjectsByUserID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to get projects"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success", "projects": projects})
}

// GetProject returns a single project
// GET /api/projects/:id
func (h *Handler) GetProject(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid project ID"})
	}

	ctx := c.Request().Context()
	project, err := h.DB.GetProjectByID(ctx, projectID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"status": "error", "error": "Project not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success", "project": project})
}

// CreateProjectAPI creates a new project
// POST /api/projects
func (h *Handler) CreateProjectAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid request"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Name is required"})
	}

	ctx := c.Request().Context()
	project, err := h.DB.CreateProject(ctx, userID, req.Name, req.Description, req.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to create project"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"status": "success", "project": project})
}

// UpdateProjectAPI updates a project
// PUT /api/projects/:id
func (h *Handler) UpdateProjectAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid project ID"})
	}

	// Verify ownership
	ctx := c.Request().Context()
	_, err = h.DB.GetProjectByID(ctx, projectID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"status": "error", "error": "Project not found"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid request"})
	}

	if err := h.DB.UpdateProject(ctx, projectID, req.Name, req.Description, req.Status); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to update project"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}

// DeleteProjectAPI deletes a project (soft delete)
// DELETE /api/projects/:id
func (h *Handler) DeleteProjectAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid project ID"})
	}

	ctx := c.Request().Context()
	_, err = h.DB.GetProjectByID(ctx, projectID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"status": "error", "error": "Project not found"})
	}

	if err := h.DB.SoftDeleteProject(ctx, projectID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to delete project"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}

// ========== TASK HANDLERS ==========

// GetTasks returns all tasks for a project
// GET /api/projects/:id/tasks
func (h *Handler) GetTasks(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid project ID"})
	}

	ctx := c.Request().Context()
	tasks, err := h.DB.GetTasksByProjectID(ctx, projectID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to get tasks"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success", "tasks": tasks})
}

// CreateTaskAPI creates a new task
// POST /api/projects/:id/tasks
func (h *Handler) CreateTaskAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid project ID"})
	}

	var req struct {
		Title        string     `json:"title"`
		Description  string     `json:"description"`
		Status       string     `json:"status"`
		Priority     string     `json:"priority"`
		DueDate      *string    `json:"due_date"`
		ParentTaskID *uuid.UUID `json:"parent_task_id"`
		DisplayOrder int        `json:"display_order"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid request"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Title is required"})
	}

	ctx := c.Request().Context()
	task, err := h.DB.CreateTask(ctx, userID, projectID, req.ParentTaskID, req.Title, req.Description, req.Status, req.Priority, nil, req.DisplayOrder)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": fmt.Sprintf("Failed to create task: %v", err)})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"status": "success", "task": task})
}

// UpdateTaskAPI updates a task
// PUT /api/tasks/:id
func (h *Handler) UpdateTaskAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}
	_ = userID

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid task ID"})
	}

	var req struct {
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		Status       string  `json:"status"`
		Priority     string  `json:"priority"`
		DueDate      *string `json:"due_date"`
		DisplayOrder int     `json:"display_order"`
		Collapsed    bool    `json:"collapsed"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid request"})
	}

	ctx := c.Request().Context()
	if err := h.DB.UpdateTask(ctx, taskID, req.Title, req.Description, req.Status, req.Priority, nil, req.DisplayOrder, req.Collapsed); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to update task"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}

// DeleteTaskAPI deletes a task (soft delete)
// DELETE /api/tasks/:id
func (h *Handler) DeleteTaskAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}
	_ = userID

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid task ID"})
	}

	ctx := c.Request().Context()
	if err := h.DB.SoftDeleteTask(ctx, taskID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to delete task"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}

// ========== TASK COMMENT HANDLERS ==========

// GetTaskComments returns all comments for a task
// GET /api/tasks/:id/comments
func (h *Handler) GetTaskComments(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid task ID"})
	}

	ctx := c.Request().Context()
	comments, err := h.DB.GetCommentsByTaskID(ctx, taskID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to get comments"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success", "comments": comments})
}

// CreateTaskCommentAPI creates a new comment
// POST /api/tasks/:id/comments
func (h *Handler) CreateTaskCommentAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid task ID"})
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid request"})
	}

	if req.Comment == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Comment is required"})
	}

	ctx := c.Request().Context()
	comment, err := h.DB.CreateTaskComment(ctx, taskID, userID, req.Comment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to create comment"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"status": "success", "comment": comment})
}

// DeleteTaskCommentAPI deletes a comment (soft delete)
// DELETE /api/comments/:id
func (h *Handler) DeleteTaskCommentAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}
	_ = userID

	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid comment ID"})
	}

	ctx := c.Request().Context()
	if err := h.DB.SoftDeleteTaskComment(ctx, commentID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to delete comment"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}

// ========== TASK ATTACHMENT HANDLERS ==========

// UploadTaskAttachment uploads an image attachment for a task
// POST /api/tasks/:id/attachments
func (h *Handler) UploadTaskAttachment(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid task ID"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Image file required"})
	}

	// Validate mime type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Only image files are allowed"})
	}

	// Validate size (10MB max)
	if file.Size > 10*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "File too large (max 10MB)"})
	}

	// Create upload directory
	uploadDir := filepath.Join("uploads", userID.String(), "tasks")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to create upload directory"})
	}

	// Generate filename
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("task_%s_%s%s", taskID.String()[:8], uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to read file"})
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to save file"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to save file"})
	}

	// Save to database
	ctx := c.Request().Context()
	attachment, err := h.DB.SaveTaskAttachment(ctx, taskID, userID, filename, "/"+filePath, file.Header.Get("Content-Type"), file.Size)
	if err != nil {
		os.Remove(filePath)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to save attachment record"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "success",
		"attachment": map[string]interface{}{
			"id":       attachment.ID,
			"taskId":   attachment.TaskID,
			"filename": attachment.Filename,
			"filePath": attachment.FilePath,
			"fileSize": attachment.FileSize,
			"mimeType": attachment.MimeType,
		},
	})
}

// DeleteTaskAttachmentAPI deletes a task attachment (soft delete)
// DELETE /api/attachments/:id
func (h *Handler) DeleteTaskAttachmentAPI(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"status": "error", "error": "Unauthorized"})
	}
	_ = userID

	attachmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "Invalid attachment ID"})
	}

	ctx := c.Request().Context()
	if err := h.DB.SoftDeleteTaskAttachment(ctx, attachmentID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": "Failed to delete attachment"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}
