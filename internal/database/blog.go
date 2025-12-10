package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetBlogPosts retrieves all blog posts for a user
func (db *DB) GetBlogPosts(ctx context.Context, userID uuid.UUID) ([]MarkdownNote, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, content, is_rtl, created_at, updated_at
		FROM markdown_notes
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []MarkdownNote
	for rows.Next() {
		var p MarkdownNote
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.IsRTL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

// SearchBlogPosts searches blog posts by title or content
func (db *DB) SearchBlogPosts(ctx context.Context, userID uuid.UUID, query string) ([]MarkdownNote, error) {
	searchPattern := "%" + query + "%"
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, content, is_rtl, created_at, updated_at
		FROM markdown_notes
		WHERE user_id = $1 AND (title ILIKE $2 OR content ILIKE $2)
		ORDER BY updated_at DESC
	`, userID, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []MarkdownNote
	for rows.Next() {
		var p MarkdownNote
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.IsRTL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

// GetBlogPost retrieves a single blog post by ID
func (db *DB) GetBlogPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*MarkdownNote, error) {
	var p MarkdownNote
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, title, content, is_rtl, created_at, updated_at
		FROM markdown_notes
		WHERE id = $1 AND user_id = $2
	`, postID, userID).Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.IsRTL, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

// CreateBlogPost creates a new blog post
func (db *DB) CreateBlogPost(ctx context.Context, userID uuid.UUID, title string) (*MarkdownNote, error) {
	var p MarkdownNote
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO markdown_notes (user_id, title, content, is_rtl)
		VALUES ($1, $2, '', true)
		RETURNING id, user_id, title, content, is_rtl, created_at, updated_at
	`, userID, title).Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.IsRTL, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// UpdateBlogPost updates an existing blog post
func (db *DB) UpdateBlogPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID, title, content string) (*MarkdownNote, error) {
	var p MarkdownNote
	err := db.Pool.QueryRow(ctx, `
		UPDATE markdown_notes
		SET title = $3, content = $4, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, title, content, is_rtl, created_at, updated_at
	`, postID, userID, title, content).Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.IsRTL, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

// DeleteBlogPost deletes a blog post
func (db *DB) DeleteBlogPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `
		DELETE FROM markdown_notes
		WHERE id = $1 AND user_id = $2
	`, postID, userID)
	return err
}
