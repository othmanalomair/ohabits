package handlers

import (
	"ohabits/internal/config"
	"ohabits/internal/database"
	"ohabits/internal/middleware"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	DB     *database.DB
	Config *config.Config
	Auth   *middleware.AuthMiddleware
}

// New creates a new Handler instance
func New(db *database.DB, cfg *config.Config, auth *middleware.AuthMiddleware) *Handler {
	return &Handler{
		DB:     db,
		Config: cfg,
		Auth:   auth,
	}
}
