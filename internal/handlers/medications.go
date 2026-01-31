package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ohabits/internal/middleware"
	"ohabits/templates/pages"
	"ohabits/templates/partials"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// medDayNames maps form index to day name
var medDayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

// MedicationsPage renders the medications management page
func (h *Handler) MedicationsPage(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.DB.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	medications, _ := h.DB.GetActiveMedications(c.Request().Context(), userID)

	return Render(c, http.StatusOK, pages.MedicationsPage(user, medications))
}

// ToggleMedication toggles a specific dose of a medication
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

	// Get dose number (1-based, 0 means reset all)
	doseNumber := 1
	if dn := c.FormValue("dose_number"); dn != "" {
		doseNumber, _ = strconv.Atoi(dn)
	}

	ctx := c.Request().Context()

	if doseNumber == 0 {
		// Reset all doses - get medication to know how many doses
		med, err := h.DB.GetMedicationByID(ctx, medID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
		}
		// Set all doses to false
		for i := 1; i <= med.TimesPerDay; i++ {
			h.DB.UpsertMedicationLog(ctx, userID, medID, date, false, i)
		}
	} else {
		// Toggle specific dose
		_, err = h.DB.ToggleMedicationLog(ctx, userID, medID, date, doseNumber)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ"})
		}
	}

	// Get updated medications
	medications, _ := h.DB.GetMedicationsForDay(ctx, userID, date)

	// Find the toggled medication
	for _, med := range medications {
		if med.ID == medID {
			allTaken := true
			for _, taken := range med.DoseTaken {
				if !taken {
					allTaken = false
					break
				}
			}
			return Render(c, http.StatusOK, partials.MedicationItem(med, date, allTaken, doseNumber))
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
	notes := c.FormValue("notes")
	icon := c.FormValue("icon")
	if icon == "" {
		icon = "pill.fill"
	}
	durationType := c.FormValue("duration_type")

	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	if durationType == "" {
		durationType = "lifetime"
	}

	// Parse times per day
	timesPerDay := 1
	if tpd := c.FormValue("times_per_day"); tpd != "" {
		timesPerDay, _ = strconv.Atoi(tpd)
		if timesPerDay < 1 {
			timesPerDay = 1
		}
	}

	// Parse scheduled days from form
	var scheduledDays []string
	allSelected := true
	for i := 0; i < 7; i++ {
		if c.FormValue("day_"+strconv.Itoa(i)) == "on" {
			scheduledDays = append(scheduledDays, medDayNames[i])
		} else {
			allSelected = false
		}
	}
	// If all days are selected or none, use empty (means daily)
	if allSelected || len(scheduledDays) == 0 {
		scheduledDays = []string{}
	}

	// Parse dates - start_date is required in DB
	var startDate, endDate *time.Time
	if sd := c.FormValue("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = &t
		}
	}
	// If no start_date provided, use today
	if startDate == nil {
		now := time.Now()
		startDate = &now
	}
	if durationType == "limited" {
		if ed := c.FormValue("end_date"); ed != "" {
			if t, err := time.Parse("2006-01-02", ed); err == nil {
				endDate = &t
			}
		}
	}

	_, err := h.DB.CreateMedication(c.Request().Context(), userID, name, dosage, scheduledDays, timesPerDay, durationType, startDate, endDate, notes, icon)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ: " + err.Error()})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"med_saved","type":"success"}}`)

	// Check if request is from medications management page
	referer := c.Request().Referer()
	if strings.Contains(referer, "/medications") {
		medications, _ := h.DB.GetActiveMedications(c.Request().Context(), userID)
		return Render(c, http.StatusOK, pages.MedicationsManageList(medications))
	}

	// Return updated list for dashboard
	date := time.Now()
	medications, _ := h.DB.GetMedicationsForDay(c.Request().Context(), userID, date)

	return Render(c, http.StatusOK, partials.MedicationsList(medications, date))
}

// UpdateMedication updates a medication
func (h *Handler) UpdateMedication(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
	}

	medIDStr := c.Param("id")
	medID, err := uuid.Parse(medIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "معرف غير صالح"})
	}

	name := c.FormValue("name")
	dosage := c.FormValue("dosage")
	notes := c.FormValue("notes")
	icon := c.FormValue("icon")
	if icon == "" {
		icon = "pill.fill"
	}
	durationType := c.FormValue("duration_type")
	isActive := c.FormValue("is_active") == "on"

	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "الاسم مطلوب"})
	}

	if durationType == "" {
		durationType = "lifetime"
	}

	// Parse times per day
	timesPerDay := 1
	if tpd := c.FormValue("times_per_day"); tpd != "" {
		timesPerDay, _ = strconv.Atoi(tpd)
		if timesPerDay < 1 {
			timesPerDay = 1
		}
	}

	// Parse scheduled days from form
	var scheduledDays []string
	allSelected := true
	for i := 0; i < 7; i++ {
		if c.FormValue("day_"+strconv.Itoa(i)) == "on" {
			scheduledDays = append(scheduledDays, medDayNames[i])
		} else {
			allSelected = false
		}
	}
	// If all days are selected or none, use empty (means daily)
	if allSelected || len(scheduledDays) == 0 {
		scheduledDays = []string{}
	}

	// Parse dates - start_date is required in DB
	var startDate, endDate *time.Time
	if sd := c.FormValue("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = &t
		}
	}
	// If no start_date provided, use today
	if startDate == nil {
		now := time.Now()
		startDate = &now
	}
	if durationType == "limited" {
		if ed := c.FormValue("end_date"); ed != "" {
			if t, err := time.Parse("2006-01-02", ed); err == nil {
				endDate = &t
			}
		}
	}

	if err := h.DB.UpdateMedication(c.Request().Context(), medID, name, dosage, scheduledDays, timesPerDay, durationType, startDate, endDate, notes, icon, isActive); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "حدث خطأ: " + err.Error()})
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"med_saved","type":"success"}}`)

	// Get updated medication and return
	med, err := h.DB.GetMedicationByID(c.Request().Context(), medID)
	if err != nil {
		// Fallback to list
		medications, _ := h.DB.GetActiveMedications(c.Request().Context(), userID)
		return Render(c, http.StatusOK, pages.MedicationsManageList(medications))
	}

	return Render(c, http.StatusOK, pages.MedicationManageItem(*med))
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

	// Check if request is from medications management page
	referer := c.Request().Referer()
	if strings.Contains(referer, "/medications") {
		medications, _ := h.DB.GetActiveMedications(c.Request().Context(), userID)
		c.Response().Header().Set("HX-Trigger", `{"showToast":{"code":"med_deleted","type":"success"}}`)
		return Render(c, http.StatusOK, pages.MedicationsManageList(medications))
	}

	// Return updated list for dashboard
	date := time.Now()
	medications, _ := h.DB.GetMedicationsForDay(c.Request().Context(), userID, date)

	return Render(c, http.StatusOK, partials.MedicationsList(medications, date))
}
