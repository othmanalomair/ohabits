package handlers

import (
	"net/http"
	"time"

	"ohabits/internal/middleware"

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

// Arabic month names for AI prompt
var arabicMonthNames = map[time.Month]string{
	time.January:   "يناير",
	time.February:  "فبراير",
	time.March:     "مارس",
	time.April:     "أبريل",
	time.May:       "مايو",
	time.June:      "يونيو",
	time.July:      "يوليو",
	time.August:    "أغسطس",
	time.September: "سبتمبر",
	time.October:   "أكتوبر",
	time.November:  "نوفمبر",
	time.December:  "ديسمبر",
}

// MonthlySummaryRequest represents the request for monthly summary
type MonthlySummaryRequest struct {
	Year  int `json:"year" form:"year"`
	Month int `json:"month" form:"month"`
}

// MonthlySummaryResponse represents the response for monthly summary
type MonthlySummaryResponse struct {
	Summary       string `json:"summary"`
	IsAIGenerated bool   `json:"is_ai_generated"`
	Error         string `json:"error,omitempty"`
}

// SaveMonthlySummaryRequest represents the request for saving monthly summary
type SaveMonthlySummaryRequest struct {
	Year    int    `json:"year" form:"year"`
	Month   int    `json:"month" form:"month"`
	Summary string `json:"summary" form:"summary"`
}

// AIGenerateMonthlySummary handles POST /api/ai/monthly-summary
func (h *Handler) AIGenerateMonthlySummary(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, MonthlySummaryResponse{
			Error: "غير مصرح",
		})
	}

	// Check if AI service is available
	if h.AI == nil {
		return c.JSON(http.StatusServiceUnavailable, MonthlySummaryResponse{
			Error: "خدمة الذكاء الاصطناعي غير متوفرة",
		})
	}

	var req MonthlySummaryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "بيانات غير صالحة",
		})
	}

	// Validate year and month
	if req.Year < 2020 || req.Year > 2100 || req.Month < 1 || req.Month > 12 {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "السنة أو الشهر غير صالح",
		})
	}

	ctx := c.Request().Context()

	// Get all notes for the month
	notesContent, err := h.DB.GetAllNotesTextForMonth(ctx, userID, req.Year, req.Month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MonthlySummaryResponse{
			Error: "فشل جلب المذكرات: " + err.Error(),
		})
	}

	if notesContent == "" {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "لا توجد مذكرات في هذا الشهر",
		})
	}

	// Generate summary with AI
	monthName := arabicMonthNames[time.Month(req.Month)]
	summary, err := h.AI.GenerateMonthlySummary(ctx, monthName, req.Year, notesContent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MonthlySummaryResponse{
			Error: "فشل توليد الملخص: " + err.Error(),
		})
	}

	// Save the summary
	_, err = h.DB.SaveMonthlySummary(ctx, userID, req.Year, req.Month, summary, true)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MonthlySummaryResponse{
			Error: "فشل حفظ الملخص: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, MonthlySummaryResponse{
		Summary:       summary,
		IsAIGenerated: true,
	})
}

// GetMonthlySummary handles GET /api/monthly-summary
func (h *Handler) GetMonthlySummary(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, MonthlySummaryResponse{
			Error: "غير مصرح",
		})
	}

	year, err := parseIntParam(c.QueryParam("year"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "السنة غير صالحة",
		})
	}

	month, err := parseIntParam(c.QueryParam("month"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "الشهر غير صالح",
		})
	}

	ctx := c.Request().Context()
	summary, err := h.DB.GetMonthlySummary(ctx, userID, year, month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MonthlySummaryResponse{
			Error: "فشل جلب الملخص: " + err.Error(),
		})
	}

	if summary == nil {
		return c.JSON(http.StatusOK, MonthlySummaryResponse{
			Summary: "",
		})
	}

	return c.JSON(http.StatusOK, MonthlySummaryResponse{
		Summary:       summary.SummaryText,
		IsAIGenerated: summary.IsAIGenerated,
	})
}

// SaveMonthlySummary handles POST /api/monthly-summary/save (manual edit)
func (h *Handler) SaveMonthlySummary(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, MonthlySummaryResponse{
			Error: "غير مصرح",
		})
	}

	var req SaveMonthlySummaryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "بيانات غير صالحة",
		})
	}

	// Validate
	if req.Year < 2020 || req.Year > 2100 || req.Month < 1 || req.Month > 12 {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "السنة أو الشهر غير صالح",
		})
	}

	if req.Summary == "" {
		return c.JSON(http.StatusBadRequest, MonthlySummaryResponse{
			Error: "الملخص مطلوب",
		})
	}

	ctx := c.Request().Context()

	// Save with is_ai_generated = false (manual edit)
	_, err := h.DB.SaveMonthlySummary(ctx, userID, req.Year, req.Month, req.Summary, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MonthlySummaryResponse{
			Error: "فشل حفظ الملخص: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, MonthlySummaryResponse{
		Summary:       req.Summary,
		IsAIGenerated: false,
	})
}
