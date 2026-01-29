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
