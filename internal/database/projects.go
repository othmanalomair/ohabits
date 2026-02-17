package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ========== PROJECTS ==========

// GetProjectsByUserID retrieves all projects for a user (including deleted for sync)
func (db *DB) GetProjectsByUserID(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, COALESCE(description, '') as description, 
		       COALESCE(status, 'active') as status, COALESCE(is_deleted, false) as is_deleted,
		       created_at, updated_at
		FROM projects WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetActiveProjectsByUserID retrieves only active (non-deleted) projects
func (db *DB) GetActiveProjectsByUserID(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, COALESCE(description, '') as description,
		       COALESCE(status, 'active') as status, COALESCE(is_deleted, false) as is_deleted,
		       created_at, updated_at
		FROM projects WHERE user_id = $1 AND COALESCE(is_deleted, false) = false
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetProjectByID retrieves a single project
func (db *DB) GetProjectByID(ctx context.Context, projectID, userID uuid.UUID) (*Project, error) {
	var p Project
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, COALESCE(description, '') as description,
		       COALESCE(status, 'active') as status, COALESCE(is_deleted, false) as is_deleted,
		       created_at, updated_at
		FROM projects WHERE id = $1 AND user_id = $2
	`, projectID, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProject creates a new project
func (db *DB) CreateProject(ctx context.Context, userID uuid.UUID, name, description, status string) (*Project, error) {
	if status == "" {
		status = "active"
	}
	var p Project
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO projects (user_id, name, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, COALESCE(description, ''), status, COALESCE(is_deleted, false), created_at, updated_at
	`, userID, name, description, status).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateProject updates a project
func (db *DB) UpdateProject(ctx context.Context, projectID uuid.UUID, name, description, status string) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE projects SET name = $2, description = $3, status = $4, updated_at = NOW()
		WHERE id = $1
	`, projectID, name, description, status)
	return err
}

// SoftDeleteProject marks a project as deleted
func (db *DB) SoftDeleteProject(ctx context.Context, projectID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE projects SET is_deleted = true, updated_at = NOW()
		WHERE id = $1
	`, projectID)
	return err
}

// getProjectsUpdatedSince retrieves projects updated since a timestamp
func (db *DB) getProjectsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Project, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, COALESCE(description, '') as description,
		       COALESCE(status, 'active') as status, COALESCE(is_deleted, false) as is_deleted,
		       created_at, updated_at
		FROM projects WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// ========== TASKS ==========

// GetTasksByProjectID retrieves all tasks for a project
func (db *DB) GetTasksByProjectID(ctx context.Context, projectID, userID uuid.UUID) ([]Task, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, project_id, parent_task_id, title, COALESCE(description, '') as description,
		       status, priority, due_date, COALESCE(completed, false) as completed,
		       COALESCE(display_order, 0) as display_order, COALESCE(collapsed, false) as collapsed,
		       COALESCE(is_deleted, false) as is_deleted, created_at, updated_at
		FROM tasks WHERE project_id = $1 AND user_id = $2 AND COALESCE(is_deleted, false) = false
		ORDER BY display_order, created_at
	`, projectID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.Completed,
			&t.DisplayOrder, &t.Collapsed, &t.IsDeleted, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetAllTasksByUserID retrieves all tasks for a user (for sync)
func (db *DB) GetAllTasksByUserID(ctx context.Context, userID uuid.UUID) ([]Task, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, project_id, parent_task_id, title, COALESCE(description, '') as description,
		       status, priority, due_date, COALESCE(completed, false) as completed,
		       COALESCE(display_order, 0) as display_order, COALESCE(collapsed, false) as collapsed,
		       COALESCE(is_deleted, false) as is_deleted, created_at, updated_at
		FROM tasks WHERE user_id = $1
		ORDER BY display_order, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.Completed,
			&t.DisplayOrder, &t.Collapsed, &t.IsDeleted, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// CreateTask creates a new task
func (db *DB) CreateTask(ctx context.Context, userID, projectID uuid.UUID, parentTaskID *uuid.UUID, title, description, status, priority string, dueDate *time.Time, displayOrder int) (*Task, error) {
	if status == "" {
		status = "Not Started"
	}
	if priority == "" {
		priority = "None"
	}
	var t Task
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO tasks (user_id, project_id, parent_task_id, title, description, status, priority, due_date, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, project_id, parent_task_id, title, COALESCE(description, ''), status, priority, due_date, completed, display_order, collapsed, COALESCE(is_deleted, false), created_at, updated_at
	`, userID, projectID, parentTaskID, title, description, status, priority, dueDate, displayOrder).Scan(
		&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.DisplayOrder, &t.Collapsed, &t.IsDeleted, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTask updates a task
func (db *DB) UpdateTask(ctx context.Context, taskID uuid.UUID, title, description, status, priority string, dueDate *time.Time, displayOrder int, collapsed bool) error {
	completed := status == "Completed"
	_, err := db.Pool.Exec(ctx, `
		UPDATE tasks SET title = $2, description = $3, status = $4, priority = $5,
		       due_date = $6, display_order = $7, collapsed = $8, completed = $9, updated_at = NOW()
		WHERE id = $1
	`, taskID, title, description, status, priority, dueDate, displayOrder, collapsed, completed)
	return err
}

// SoftDeleteTask marks a task as deleted
func (db *DB) SoftDeleteTask(ctx context.Context, taskID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE tasks SET is_deleted = true, updated_at = NOW()
		WHERE id = $1
	`, taskID)
	return err
}

// getTasksUpdatedSince retrieves tasks updated since a timestamp
func (db *DB) getTasksUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]Task, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, project_id, parent_task_id, title, COALESCE(description, '') as description,
		       status, priority, due_date, COALESCE(completed, false) as completed,
		       COALESCE(display_order, 0) as display_order, COALESCE(collapsed, false) as collapsed,
		       COALESCE(is_deleted, false) as is_deleted, created_at, updated_at
		FROM tasks WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.Completed,
			&t.DisplayOrder, &t.Collapsed, &t.IsDeleted, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ========== TASK COMMENTS ==========

// GetCommentsByTaskID retrieves all comments for a task
func (db *DB) GetCommentsByTaskID(ctx context.Context, taskID, userID uuid.UUID) ([]TaskComment, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, task_id, user_id, comment, COALESCE(is_deleted, false) as is_deleted, created_at, updated_at
		FROM task_comments WHERE task_id = $1 AND user_id = $2 AND COALESCE(is_deleted, false) = false
		ORDER BY created_at
	`, taskID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []TaskComment
	for rows.Next() {
		var c TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Comment, &c.IsDeleted, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// GetAllCommentsByUserID retrieves all comments for a user (for sync)
func (db *DB) GetAllCommentsByUserID(ctx context.Context, userID uuid.UUID) ([]TaskComment, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, task_id, user_id, comment, COALESCE(is_deleted, false) as is_deleted, created_at, updated_at
		FROM task_comments WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []TaskComment
	for rows.Next() {
		var c TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Comment, &c.IsDeleted, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// CreateTaskComment creates a new comment
func (db *DB) CreateTaskComment(ctx context.Context, taskID, userID uuid.UUID, comment string) (*TaskComment, error) {
	var c TaskComment
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO task_comments (task_id, user_id, comment)
		VALUES ($1, $2, $3)
		RETURNING id, task_id, user_id, comment, COALESCE(is_deleted, false), created_at, updated_at
	`, taskID, userID, comment).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Comment, &c.IsDeleted, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateTaskComment updates a comment
func (db *DB) UpdateTaskComment(ctx context.Context, commentID uuid.UUID, comment string) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE task_comments SET comment = $2, updated_at = NOW()
		WHERE id = $1
	`, commentID, comment)
	return err
}

// SoftDeleteTaskComment marks a comment as deleted
func (db *DB) SoftDeleteTaskComment(ctx context.Context, commentID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE task_comments SET is_deleted = true, updated_at = NOW()
		WHERE id = $1
	`, commentID)
	return err
}

// getCommentsUpdatedSince retrieves comments updated since a timestamp
func (db *DB) getCommentsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]TaskComment, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, task_id, user_id, comment, COALESCE(is_deleted, false) as is_deleted, created_at, updated_at
		FROM task_comments WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []TaskComment
	for rows.Next() {
		var c TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Comment, &c.IsDeleted, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// ========== TASK ATTACHMENTS ==========

// GetAttachmentsByTaskID retrieves all attachments for a task
func (db *DB) GetAttachmentsByTaskID(ctx context.Context, taskID, userID uuid.UUID) ([]TaskAttachment, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, task_id, user_id, filename, file_path, file_size, mime_type, COALESCE(is_deleted, false) as is_deleted, created_at
		FROM task_attachments WHERE task_id = $1 AND user_id = $2 AND COALESCE(is_deleted, false) = false
		ORDER BY created_at
	`, taskID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []TaskAttachment
	for rows.Next() {
		var a TaskAttachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.FilePath, &a.FileSize, &a.MimeType, &a.IsDeleted, &a.CreatedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}

// GetAllAttachmentsByUserID retrieves all attachments for a user (for sync)
func (db *DB) GetAllAttachmentsByUserID(ctx context.Context, userID uuid.UUID) ([]TaskAttachment, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, task_id, user_id, filename, file_path, file_size, mime_type, COALESCE(is_deleted, false) as is_deleted, created_at
		FROM task_attachments WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []TaskAttachment
	for rows.Next() {
		var a TaskAttachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.FilePath, &a.FileSize, &a.MimeType, &a.IsDeleted, &a.CreatedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}

// SaveTaskAttachment saves a task attachment record
func (db *DB) SaveTaskAttachment(ctx context.Context, taskID, userID uuid.UUID, filename, filePath, mimeType string, fileSize int64) (*TaskAttachment, error) {
	var a TaskAttachment
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO task_attachments (task_id, user_id, filename, file_path, file_size, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, task_id, user_id, filename, file_path, file_size, mime_type, COALESCE(is_deleted, false), created_at
	`, taskID, userID, filename, filePath, fileSize, mimeType).Scan(
		&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.FilePath, &a.FileSize, &a.MimeType, &a.IsDeleted, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// SoftDeleteTaskAttachment marks an attachment as deleted
func (db *DB) SoftDeleteTaskAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE task_attachments SET is_deleted = true
		WHERE id = $1
	`, attachmentID)
	return err
}

// getAttachmentsUpdatedSince - for task attachments we use created_at since they don't have updated_at
func (db *DB) getAttachmentsCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]TaskAttachment, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, task_id, user_id, filename, file_path, file_size, mime_type, COALESCE(is_deleted, false) as is_deleted, created_at
		FROM task_attachments WHERE user_id = $1 AND created_at > $2
		ORDER BY created_at
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []TaskAttachment
	for rows.Next() {
		var a TaskAttachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.FilePath, &a.FileSize, &a.MimeType, &a.IsDeleted, &a.CreatedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}
