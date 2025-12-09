package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidPassword = errors.New("invalid password")
)

// CreateUser creates a new user
func (db *DB) CreateUser(ctx context.Context, email, password, displayName string) (*User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user User
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, display_name, avatar_url, created_at, updated_at
	`, email, string(hashedPassword), displayName).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(ctx, `
		SELECT id, email, password, display_name, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := db.Pool.QueryRow(ctx, `
		SELECT id, email, password, display_name, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// AuthenticateUser checks email and password
func (db *DB) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	user, err := db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// UpdateUserProfile updates user display name and avatar
func (db *DB) UpdateUserProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL string) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE users SET display_name = $2, avatar_url = $3
		WHERE id = $1
	`, userID, displayName, avatarURL)
	return err
}
