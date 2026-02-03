package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SaveDailyImage saves a new image record
func (db *DB) SaveDailyImage(ctx context.Context, userID uuid.UUID, date time.Time, originalPath, thumbnailPath, filename, mimeType string, sizeBytes int) (*DailyImage, error) {
	dateStr := date.Format("2006-01-02")

	var img DailyImage
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO daily_images (user_id, date, original_path, thumbnail_path, filename, mime_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, date, original_path, thumbnail_path, filename, mime_type, size_bytes, created_at
	`, userID, dateStr, originalPath, thumbnailPath, filename, mimeType, sizeBytes).Scan(
		&img.ID, &img.UserID, &img.Date, &img.OriginalPath, &img.ThumbnailPath,
		&img.Filename, &img.MimeType, &img.SizeBytes, &img.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &img, nil
}

// GetImagesForDay retrieves all non-deleted images for a specific day
func (db *DB) GetImagesForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]DailyImage, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, date, original_path, thumbnail_path, filename, mime_type, size_bytes, created_at
		FROM daily_images
		WHERE user_id = $1 AND date = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []DailyImage
	for rows.Next() {
		var img DailyImage
		if err := rows.Scan(&img.ID, &img.UserID, &img.Date, &img.OriginalPath, &img.ThumbnailPath,
			&img.Filename, &img.MimeType, &img.SizeBytes, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()
}

// DeleteDailyImage soft-deletes an image record by setting deleted_at
func (db *DB) DeleteDailyImage(ctx context.Context, imageID, userID uuid.UUID) (*DailyImage, error) {
	var img DailyImage
	err := db.Pool.QueryRow(ctx, `
		UPDATE daily_images
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, date, original_path, thumbnail_path, filename, mime_type, size_bytes, created_at, deleted_at
	`, imageID, userID).Scan(
		&img.ID, &img.UserID, &img.Date, &img.OriginalPath, &img.ThumbnailPath,
		&img.Filename, &img.MimeType, &img.SizeBytes, &img.CreatedAt, &img.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &img, nil
}

// SaveBlogImage saves a new blog image record
func (db *DB) SaveBlogImage(ctx context.Context, userID, markdownNoteID uuid.UUID, originalPath, thumbnailPath, filename, mimeType string, sizeBytes int, positionMarker string) (*BlogImage, error) {
	var img BlogImage
	var thumbPtr *string
	if thumbnailPath != "" {
		thumbPtr = &thumbnailPath
	}

	err := db.Pool.QueryRow(ctx, `
		INSERT INTO blog_images (user_id, markdown_note_id, original_path, thumbnail_path, filename, mime_type, size_bytes, position_marker)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, markdown_note_id, original_path, thumbnail_path, filename, mime_type, size_bytes, position_marker, is_deleted, created_at, updated_at
	`, userID, markdownNoteID, originalPath, thumbPtr, filename, mimeType, sizeBytes, positionMarker).Scan(
		&img.ID, &img.UserID, &img.MarkdownNoteID, &img.OriginalPath, &img.ThumbnailPath,
		&img.Filename, &img.MimeType, &img.SizeBytes, &img.PositionMarker, &img.IsDeleted, &img.CreatedAt, &img.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &img, nil
}

// GetBlogImagesForNote retrieves all non-deleted images for a specific markdown note
func (db *DB) GetBlogImagesForNote(ctx context.Context, userID, markdownNoteID uuid.UUID) ([]BlogImage, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, markdown_note_id, original_path, thumbnail_path, filename, mime_type, size_bytes, position_marker, is_deleted, created_at, updated_at
		FROM blog_images
		WHERE user_id = $1 AND markdown_note_id = $2 AND is_deleted = false
		ORDER BY created_at ASC
	`, userID, markdownNoteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []BlogImage
	for rows.Next() {
		var img BlogImage
		if err := rows.Scan(&img.ID, &img.UserID, &img.MarkdownNoteID, &img.OriginalPath, &img.ThumbnailPath,
			&img.Filename, &img.MimeType, &img.SizeBytes, &img.PositionMarker, &img.IsDeleted, &img.CreatedAt, &img.UpdatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()
}

// DeleteBlogImage soft-deletes a blog image record
func (db *DB) DeleteBlogImage(ctx context.Context, imageID, userID uuid.UUID) (*BlogImage, error) {
	var img BlogImage
	err := db.Pool.QueryRow(ctx, `
		UPDATE blog_images
		SET is_deleted = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_deleted = false
		RETURNING id, user_id, markdown_note_id, original_path, thumbnail_path, filename, mime_type, size_bytes, position_marker, is_deleted, created_at, updated_at
	`, imageID, userID).Scan(
		&img.ID, &img.UserID, &img.MarkdownNoteID, &img.OriginalPath, &img.ThumbnailPath,
		&img.Filename, &img.MimeType, &img.SizeBytes, &img.PositionMarker, &img.IsDeleted, &img.CreatedAt, &img.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &img, nil
}

// GetAllBlogImages retrieves all blog images for a user (for sync)
func (db *DB) GetAllBlogImages(ctx context.Context, userID uuid.UUID) ([]BlogImage, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, markdown_note_id, original_path, thumbnail_path, filename, mime_type, size_bytes, position_marker, is_deleted, created_at, updated_at
		FROM blog_images
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []BlogImage
	for rows.Next() {
		var img BlogImage
		if err := rows.Scan(&img.ID, &img.UserID, &img.MarkdownNoteID, &img.OriginalPath, &img.ThumbnailPath,
			&img.Filename, &img.MimeType, &img.SizeBytes, &img.PositionMarker, &img.IsDeleted, &img.CreatedAt, &img.UpdatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()
}

// GetBlogImagesUpdatedSince retrieves blog images updated since a timestamp
func (db *DB) GetBlogImagesUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]BlogImage, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, markdown_note_id, original_path, thumbnail_path, filename, mime_type, size_bytes, position_marker, is_deleted, created_at, updated_at
		FROM blog_images
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at ASC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []BlogImage
	for rows.Next() {
		var img BlogImage
		if err := rows.Scan(&img.ID, &img.UserID, &img.MarkdownNoteID, &img.OriginalPath, &img.ThumbnailPath,
			&img.Filename, &img.MimeType, &img.SizeBytes, &img.PositionMarker, &img.IsDeleted, &img.CreatedAt, &img.UpdatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()
}
