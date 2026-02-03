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

	// Create handlers (AIService is created inside)
	h := handlers.New(db, cfg, auth)

	// Log AI service status
	if h.AIService.IsConfigured() {
		log.Println("✅ خدمة AI متاحة (OpenRouter)")
	} else {
		log.Println("⚠️  خدمة AI غير متاحة - تأكد من تعيين OPENROUTER_API_KEY")
	}

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

	// Apple Sign-In (public - for native app authentication)
	e.POST("/api/auth/apple", h.AppleSignIn)
	e.POST("/api/auth/login", h.APILogin)

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

	// AI API endpoints (OpenRouter - for blog assistant)
	aiHandler := handlers.NewAIHandler(h.AIService)
	protected.POST("/api/ai/suggest-titles", aiHandler.SuggestTitles)
	protected.POST("/api/ai/format-markdown", aiHandler.FormatMarkdown)
	protected.POST("/api/ai/custom-prompt", aiHandler.CustomPrompt)

	// Calendar Events (الرزنامة)
	protected.GET("/calendar", h.CalendarPage)
	protected.POST("/calendar", h.CreateCalendarEvent)
	protected.PUT("/calendar/:id", h.UpdateCalendarEvent)
	protected.DELETE("/calendar/:id", h.DeleteCalendarEvent)

	// Sync API (للتطبيق الأصلي iOS/macOS)
	protected.GET("/api/sync/all", h.SyncAll)
	protected.POST("/api/sync/push", h.SyncPush)
	protected.POST("/api/sync/changes", h.SyncChanges)
	protected.GET("/api/sync/status", h.SyncStatus)

	// Profile API (للتطبيق الأصلي)
	protected.GET("/api/user/profile", h.GetProfileAPI)
	protected.PUT("/api/user/profile", h.UpdateProfileAPI)
	protected.POST("/api/user/profile/image", h.UploadProfileImageAPI)
	protected.DELETE("/api/user/profile/image", h.DeleteProfileImageAPI)

	// User API (للتطبيق الأصلي)
	protected.GET("/api/user/info", h.UserInfo)
	protected.GET("/api/auth/validate", h.ValidateToken)
	protected.POST("/api/auth/refresh", h.RefreshToken)

	// Images API (للتطبيق الأصلي)
	protected.POST("/api/images", h.UploadImages)
	protected.DELETE("/api/images/:id", h.DeleteImage)

	// Blog images API
	protected.POST("/api/blog/images", h.UploadBlogImageAPI)
	protected.DELETE("/api/blog/images/:id", h.DeleteBlogImageAPI)

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
