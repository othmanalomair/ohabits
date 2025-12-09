package handlers

import (
	"net/http"
	"time"

	"ohabits/internal/middleware"
	"ohabits/templates/partials"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ToggleMedication toggles medication taken status
func (h *Handler) ToggleMedication(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	medIDStr := c.Param("id")
	medID, err := uuid.Parse(medIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	taken, err := h.DB.ToggleMedicationLog(c.Request().Context(), userID, medID, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Get updated medications
	medications, _ := h.DB.GetMedicationsForDay(c.Request().Context(), userID, date)

	// Find the toggled medication
	for _, med := range medications {
		if med.ID == medID {
			return Render(c, http.StatusOK, partials.MedicationItem(med, date, taken))
		}
	}

	return c.NoContent(http.StatusOK)
}

// CreateMedication creates a new medication
func (h *Handler) CreateMedication(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	name := c.FormValue("name")
	dosage := c.FormValue("dosage")

	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	// For simplicity, create with default values
	scheduledDays := []int{} // Empty means every day
	timesPerDay := 1
	durationType := "lifetime"

	_, err := h.DB.CreateMedication(c.Request().Context(), userID, name, dosage, scheduledDays, timesPerDay, durationType, nil, nil, "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Return updated list
	date := time.Now()
	medications, _ := h.DB.GetMedicationsForDay(c.Request().Context(), userID, date)

	return Render(c, http.StatusOK, partials.MedicationsList(medications, date))
}

// DeleteMedication deletes a medication
func (h *Handler) DeleteMedication(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	medIDStr := c.Param("id")
	medID, err := uuid.Parse(medIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	if err := h.DB.DeleteMedication(c.Request().Context(), medID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
	}

	// Return updated list
	date := time.Now()
	medications, _ := h.DB.GetMedicationsForDay(c.Request().Context(), userID, date)

	return Render(c, http.StatusOK, partials.MedicationsList(medications, date))
}
