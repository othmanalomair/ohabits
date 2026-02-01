package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GetUserSettings retrieves user settings for a user
func (db *DB) GetUserSettings(ctx context.Context, userID uuid.UUID) (*UserSettings, error) {
	var settings UserSettings
	var configsJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, section_configs, is_deleted, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1 AND is_deleted = false
	`, userID).Scan(&settings.ID, &settings.UserID, &configsJSON, &settings.IsDeleted, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		// Return nil if not found (not an error - user just has no settings yet)
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	// Parse the JSON config
	if err := json.Unmarshal(configsJSON, &settings.SectionConfigs); err != nil {
		return nil, err
	}

	return &settings, nil
}

// getUserSettingsUpdatedSince retrieves user settings if updated since timestamp
func (db *DB) getUserSettingsUpdatedSince(ctx context.Context, userID uuid.UUID, since time.Time) (*UserSettings, error) {
	var settings UserSettings
	var configsJSON []byte

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, section_configs, is_deleted, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1 AND updated_at > $2
	`, userID, since).Scan(&settings.ID, &settings.UserID, &configsJSON, &settings.IsDeleted, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		// Return nil if not found
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	// Parse the JSON config
	if err := json.Unmarshal(configsJSON, &settings.SectionConfigs); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveUserSettings creates or updates user settings
func (db *DB) SaveUserSettings(ctx context.Context, userID uuid.UUID, configs []SectionConfig) (*UserSettings, error) {
	configsJSON, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}

	var settings UserSettings
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO user_settings (user_id, section_configs, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) 
		DO UPDATE SET section_configs = $2, updated_at = NOW()
		RETURNING id, user_id, section_configs, is_deleted, created_at, updated_at
	`, userID, configsJSON).Scan(&settings.ID, &settings.UserID, &configsJSON, &settings.IsDeleted, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Parse back the JSON config
	if err := json.Unmarshal(configsJSON, &settings.SectionConfigs); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SyncPushUserSettings handles syncing user settings from the client
func (db *DB) SyncPushUserSettings(ctx context.Context, userID uuid.UUID, serverID *string, isDeleted bool, data json.RawMessage) (string, error) {
	if isDeleted {
		// Mark settings as deleted if they exist
		if serverID != nil {
			id, err := uuid.Parse(*serverID)
			if err != nil {
				return "", err
			}
			_, err = db.Pool.Exec(ctx, `
				UPDATE user_settings SET is_deleted = true, updated_at = NOW()
				WHERE id = $1 AND user_id = $2
			`, id, userID)
			return *serverID, err
		}
		return "", nil
	}

	var settingsData struct {
		SectionConfigs []SectionConfig `json:"section_configs"`
	}
	if err := json.Unmarshal(data, &settingsData); err != nil {
		return "", err
	}

	// Always upsert based on user_id (only one settings record per user)
	settings, err := db.SaveUserSettings(ctx, userID, settingsData.SectionConfigs)
	if err != nil {
		return "", err
	}
	return settings.ID.String(), nil
}
