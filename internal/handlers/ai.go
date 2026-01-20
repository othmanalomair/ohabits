package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// FixTextRequest represents the request body for text fixing
type FixTextRequest struct {
	Text   string `json:"text" form:"text"`
	Action string `json:"action" form:"action"` // improve, fix, simplify
}

// FixTextResponse represents the response for text fixing
type FixTextResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

// GenerateTitlesRequest represents the request body for title generation
type GenerateTitlesRequest struct {
	Content string `json:"content" form:"content"`
}

// GenerateTitlesResponse represents the response for title generation
type GenerateTitlesResponse struct {
	Titles []string `json:"titles"`
	Error  string   `json:"error,omitempty"`
}

// AIFixText handles POST /api/ai/fix-text
func (h *Handler) AIFixText(c echo.Context) error {
	// Check if AI service is available
	if h.AI == nil {
		return c.JSON(http.StatusServiceUnavailable, FixTextResponse{
			Error: "خدمة الذكاء الاصطناعي غير متوفرة",
		})
	}

	var req FixTextRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, FixTextResponse{
			Error: "بيانات غير صالحة",
		})
	}

	if req.Text == "" {
		return c.JSON(http.StatusBadRequest, FixTextResponse{
			Error: "النص مطلوب",
		})
	}

	// Default action is "improve"
	if req.Action == "" {
		req.Action = "improve"
	}

	// Process the text with AI
	result, err := h.AI.FixText(c.Request().Context(), req.Text, req.Action)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, FixTextResponse{
			Error: "حدث خطأ أثناء معالجة النص: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, FixTextResponse{
		Text: result,
	})
}

// AIGenerateTitles handles POST /api/ai/generate-title
func (h *Handler) AIGenerateTitles(c echo.Context) error {
	// Check if AI service is available
	if h.AI == nil {
		return c.JSON(http.StatusServiceUnavailable, GenerateTitlesResponse{
			Error: "خدمة الذكاء الاصطناعي غير متوفرة",
		})
	}

	var req GenerateTitlesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, GenerateTitlesResponse{
			Error: "بيانات غير صالحة",
		})
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, GenerateTitlesResponse{
			Error: "المحتوى مطلوب",
		})
	}

	// Generate titles with AI
	titles, err := h.AI.GenerateTitles(c.Request().Context(), req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, GenerateTitlesResponse{
			Error: "حدث خطأ أثناء توليد العناوين: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, GenerateTitlesResponse{
		Titles: titles,
	})
}

// AIStatus handles GET /api/ai/status
func (h *Handler) AIStatus(c echo.Context) error {
	if h.AI == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"available": false,
			"message":   "خدمة AI غير مُعدّة",
		})
	}

	available := h.AI.IsAvailable()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"available": available,
		"message": func() string {
			if available {
				return "خدمة AI متاحة"
			}
			return "خدمة AI غير متاحة - تأكد من تشغيل Ollama"
		}(),
	})
}
