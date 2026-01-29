package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"ohabits/internal/database"
	"ohabits/internal/middleware"

	"github.com/labstack/echo/v4"
)

// SyncAllResponse represents the response for full data sync
type SyncAllResponse struct {
	Status            string                 `json:"status"`
	Data              *database.SyncAllData  `json:"data,omitempty"`
	LastSyncTimestamp time.Time              `json:"lastSyncTimestamp"`
	Error             string                 `json:"error,omitempty"`
}

// SyncPushRequest represents the request for pushing changes
type SyncPushRequest struct {
	Items []database.SyncPushItem `json:"items"`
}

// SyncPushResponse represents the response for push operations
type SyncPushResponse struct {
	Status  string                    `json:"status"`
	Results []database.SyncPushResult `json:"results"`
	Error   string                    `json:"error,omitempty"`
}

// SyncChangesRequest represents the request for incremental sync
type SyncChangesRequest struct {
	Since time.Time `json:"since"`
}

// SyncChangesResponse represents the response for incremental sync
type SyncChangesResponse struct {
	Status            string                     `json:"status"`
	Data              *database.SyncChangesData  `json:"data,omitempty"`
	LastSyncTimestamp time.Time                  `json:"lastSyncTimestamp"`
	Error             string                     `json:"error,omitempty"`
}

// SyncAll handles full data pull for initial sync
// GET /api/sync/all
func (h *Handler) SyncAll(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, SyncAllResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	ctx := c.Request().Context()

	data, err := h.DB.GetAllSyncData(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, SyncAllResponse{
			Status: "error",
			Error:  "Failed to retrieve sync data",
		})
	}

	return c.JSON(http.StatusOK, SyncAllResponse{
		Status:            "success",
		Data:              data,
		LastSyncTimestamp: data.LastSyncTimestamp,
	})
}

// SyncPush handles batch push of changes from native app
// POST /api/sync/push
func (h *Handler) SyncPush(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, SyncPushResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	// Log raw body for debugging
	body, _ := io.ReadAll(c.Request().Body)
	log.Printf("Raw push body: %s", string(body))
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	var req SyncPushRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, SyncPushResponse{
			Status: "error",
			Error:  "Invalid request format",
		})
	}

	if len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, SyncPushResponse{
			Status: "error",
			Error:  "No items to sync",
		})
	}

	ctx := c.Request().Context()
	results := make([]database.SyncPushResult, 0, len(req.Items))

	for _, item := range req.Items {
		result := database.SyncPushResult{
			LocalID: item.LocalID,
			Success: false,
		}

		var serverID string
		var err error

		switch item.Type {
		case "habit":
			serverID, err = h.DB.SyncPushHabit(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "habitCompletion":
			serverID, err = h.DB.SyncPushHabitCompletion(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "medication":
			serverID, err = h.DB.SyncPushMedication(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "medicationLog":
			serverID, err = h.DB.SyncPushMedicationLog(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "todo":
			if item.ServerID != nil { log.Printf("Handler: todo serverID=%s", *item.ServerID) } else { log.Printf("Handler: todo serverID=nil") }
			serverID, err = h.DB.SyncPushTodo(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "note":
			serverID, err = h.DB.SyncPushNote(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "mood":
			serverID, err = h.DB.SyncPushMood(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "event":
			serverID, err = h.DB.SyncPushEvent(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "workout":
			serverID, err = h.DB.SyncPushWorkout(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "workoutLog":
			serverID, err = h.DB.SyncPushWorkoutLog(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		case "markdownNote":
			serverID, err = h.DB.SyncPushMarkdownNote(ctx, userID, item.ServerID, item.IsDeleted, item.Data)
		default:
			result.Error = "Unknown item type: " + item.Type
			results = append(results, result)
			continue
		}

		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
			result.ServerID = serverID
		}

		results = append(results, result)
	}

	return c.JSON(http.StatusOK, SyncPushResponse{
		Status:  "success",
		Results: results,
	})
}

// SyncChanges handles incremental pull since timestamp
// POST /api/sync/changes
func (h *Handler) SyncChanges(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, SyncChangesResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
	}

	var req SyncChangesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, SyncChangesResponse{
			Status: "error",
			Error:  "Invalid request format",
		})
	}

	// Default to beginning of time if not specified
	if req.Since.IsZero() {
		req.Since = time.Unix(0, 0)
	}

	ctx := c.Request().Context()

	data, err := h.DB.GetSyncChangesSince(ctx, userID, req.Since)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, SyncChangesResponse{
			Status: "error",
			Error:  "Failed to retrieve sync changes",
		})
	}

	return c.JSON(http.StatusOK, SyncChangesResponse{
		Status:            "success",
		Data:              data,
		LastSyncTimestamp: data.LastSyncTimestamp,
	})
}

// SyncStatus returns the current sync status and server timestamp
// GET /api/sync/status
func (h *Handler) SyncStatus(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "Unauthorized",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":          "success",
		"serverTimestamp": time.Now(),
		"userId":          userID.String(),
	})
}

// UserInfo returns user information for the native app
// GET /api/user/info
func (h *Handler) UserInfo(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "Unauthorized",
		})
	}

	ctx := c.Request().Context()

	user, err := h.DB.GetUserByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  "Failed to get user info",
		})
	}

	response := map[string]interface{}{
		"status": "success",
		"user": map[string]interface{}{
			"id":          user.ID.String(),
			"email":       user.Email,
			"displayName": user.DisplayName,
			"avatarUrl":   user.GetAvatarURL(),
			"hasAppleId":  user.AppleID != nil,
			"createdAt":   user.CreatedAt,
			"updatedAt":   user.UpdatedAt,
		},
	}

	return c.JSON(http.StatusOK, response)
}

// ValidateToken validates the current JWT token and returns user info
// GET /api/auth/validate
func (h *Handler) ValidateToken(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"valid": false,
			"error": "Invalid or expired token",
		})
	}

	ctx := c.Request().Context()

	user, err := h.DB.GetUserByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"valid": false,
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"valid": true,
		"user": map[string]interface{}{
			"id":          user.ID.String(),
			"email":       user.Email,
			"displayName": user.DisplayName,
		},
	})
}

// RefreshToken generates a new JWT token for the current user
// POST /api/auth/refresh
func (h *Handler) RefreshToken(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "Unauthorized",
		})
	}

	ctx := c.Request().Context()

	user, err := h.DB.GetUserByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "User not found",
		})
	}

	// Generate new token
	token, err := h.Auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  "Failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"token":  token,
		"user": map[string]interface{}{
			"id":          user.ID.String(),
			"email":       user.Email,
			"displayName": user.DisplayName,
		},
	})
}

// Helper to create a pointer to a string (used for serverID)
func stringPtr(s string) *string {
	return &s
}

// Helper to unmarshal raw JSON to a specific type
func unmarshalTo[T any](data json.RawMessage) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
