package handlers

import (
	"net/http"

	"ohabits/internal/services"

	"github.com/labstack/echo/v4"
)

// AIHandler handles AI-related API endpoints using OpenRouter
type AIHandler struct {
	AIService *services.AIService
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService *services.AIService) *AIHandler {
	return &AIHandler{
		AIService: aiService,
	}
}

// SuggestTitles handles POST /api/ai/suggest-titles
func (h *AIHandler) SuggestTitles(c echo.Context) error {
	if !h.AIService.IsConfigured() {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "error",
			"error":  "AI service not configured",
		})
	}

	var req services.SuggestTitlesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Content is required",
		})
	}

	resp, err := h.AIService.SuggestTitles(req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"titles": resp.Titles,
	})
}

// FormatMarkdown handles POST /api/ai/format-markdown
func (h *AIHandler) FormatMarkdown(c echo.Context) error {
	if !h.AIService.IsConfigured() {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "error",
			"error":  "AI service not configured",
		})
	}

	var req services.FormatMarkdownRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Content is required",
		})
	}

	resp, err := h.AIService.FormatMarkdown(req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":           "success",
		"formattedContent": resp.FormattedContent,
	})
}

// CustomPrompt handles POST /api/ai/custom-prompt
func (h *AIHandler) CustomPrompt(c echo.Context) error {
	if !h.AIService.IsConfigured() {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "error",
			"error":  "AI service not configured",
		})
	}

	var req services.CustomPromptRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Content is required",
		})
	}

	if req.Prompt == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Prompt is required",
		})
	}

	resp, err := h.AIService.CustomPrompt(req.Content, req.Prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"result": resp.Result,
	})
}
