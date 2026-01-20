package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ohabits/internal/config"
	"ohabits/internal/database"
	"ohabits/internal/handlers"
	"ohabits/internal/middleware"
	"ohabits/internal/services/ai"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("فشل الاتصال بقاعدة البيانات: %v", err)
	}
	defer db.Close()

	log.Println("✅ تم الاتصال بقاعدة البيانات")

	// Create auth middleware
	auth := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Create AI service
	aiService := ai.New(cfg.OllamaURL, cfg.AIModel)
	if aiService.IsAvailable() {
		log.Println("✅ خدمة AI متاحة (Ollama)")
	} else {
		log.Println("⚠️  خدمة AI غير متاحة - تأكد من تشغيل Ollama")
	}

	// Create handlers
	h := handlers.New(db, cfg, auth, aiService)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.Gzip())

	// Rate Limiting - حماية من هجمات Brute Force
	e.Use(echomw.RateLimiter(echomw.NewRateLimiterMemoryStore(20))) // 20 request/second عام

	// Security Headers - حماية من XSS وهجمات أخرى
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: blob:;",
	}))

	// Static files
	e.Static("/static", "static")
	e.Static("/uploads", "uploads")
	e.File("/favicon.ico", "static/images/icons/icon-192x192.png")

	// Rate limiter صارم للـ Login (5 محاولات بالدقيقة)
	loginLimiter := echomw.RateLimiter(echomw.NewRateLimiterMemoryStoreWithConfig(
		echomw.RateLimiterMemoryStoreConfig{Rate: 5, Burst: 5, ExpiresIn: time.Minute},
	))

	// Public routes
	e.GET("/login", h.LoginPage)
	e.POST("/login", h.Login, loginLimiter)
	e.GET("/signup", h.SignupPage)
	e.POST("/signup", h.Signup)
	e.GET("/logout", h.Logout)

	// Protected routes
	protected := e.Group("")
	protected.Use(auth.RequireAuth)

	// Dashboard
	protected.GET("/", h.Dashboard)

	// Habits
	protected.GET("/habits", h.HabitsPage)
	protected.POST("/habits", h.CreateHabit)
	protected.PUT("/habits/:id", h.UpdateHabit)
	protected.POST("/habits/:id/toggle", h.ToggleHabit)
	protected.DELETE("/habits/:id", h.DeleteHabit)

	// Medications
	protected.GET("/medications", h.MedicationsPage)
	protected.POST("/medications", h.CreateMedication)
	protected.PUT("/medications/:id", h.UpdateMedication)
	protected.POST("/medications/:id/toggle", h.ToggleMedication)
	protected.DELETE("/medications/:id", h.DeleteMedication)

	// Todos
	protected.POST("/todos", h.CreateTodo)
	protected.POST("/todos/:id/toggle", h.ToggleTodo)
	protected.DELETE("/todos/:id", h.DeleteTodo)

	// Notes & Mood
	protected.POST("/notes", h.SaveNote)
	protected.POST("/mood", h.SaveMood)
	protected.GET("/daily-notes", h.DailyNotesPage)
	protected.GET("/notes/search", h.SearchNotes)

	// Images
	protected.POST("/images", h.UploadImages)
	protected.DELETE("/images/:id", h.DeleteImage)

	// Workouts
	protected.GET("/workouts", h.WorkoutsPage)
	protected.POST("/workouts", h.CreateWorkout)
	protected.PUT("/workouts/:id", h.UpdateWorkout)
	protected.DELETE("/workouts/:id", h.DeleteWorkout)
	protected.POST("/workouts/reorder", h.ReorderWorkouts)
	protected.POST("/workout-log", h.SaveWorkoutLog)

	// Profile
	protected.GET("/profile", h.ProfilePage)
	protected.POST("/profile/info", h.UpdateProfileInfo)
	protected.POST("/profile/password", h.UpdateProfilePassword)
	protected.POST("/profile/avatar", h.UpdateProfileAvatar)

	// Blog
	protected.GET("/blog", h.BlogPage)
	protected.GET("/blog/search", h.BlogSearch)
	protected.GET("/blog/new", h.BlogNewPage)
	protected.GET("/blog/:id", h.BlogViewPage)
	protected.GET("/blog/:id/edit", h.BlogEditPage)
	protected.POST("/blog/:id/save", h.BlogSave)
	protected.DELETE("/blog/:id", h.BlogDelete)
	protected.POST("/blog/upload-image", h.BlogUploadImage)

	// Calendar Events (الرزنامة)
	protected.GET("/calendar", h.CalendarPage)
	protected.POST("/calendar", h.CreateCalendarEvent)
	protected.PUT("/calendar/:id", h.UpdateCalendarEvent)
	protected.DELETE("/calendar/:id", h.DeleteCalendarEvent)

	// AI (الذكاء الاصطناعي)
	protected.GET("/api/ai/status", h.AIStatus)
	protected.POST("/api/ai/fix-text", h.AIFixText)
	protected.POST("/api/ai/generate-title", h.AIGenerateTitles)
	protected.POST("/api/ai/monthly-summary", h.AIGenerateMonthlySummary)

	// Monthly Summary (ملخص الشهر)
	protected.GET("/api/monthly-summary", h.GetMonthlySummary)
	protected.POST("/api/monthly-summary/save", h.SaveMonthlySummary)

	// Start server
	go func() {
		addr := ":" + cfg.Port
		log.Printf("🚀 تشغيل السيرفر على http://localhost%s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("فشل تشغيل السيرفر: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ إيقاف السيرفر...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("فشل إيقاف السيرفر: %v", err)
	}

	log.Println("👋 تم إيقاف السيرفر")
}
