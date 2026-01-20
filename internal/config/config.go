package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Env         string
	OllamaURL   string
	AIModel     string
}

func Load() *Config {
	// Load .env file if exists
	godotenv.Load()

	env := getEnv("ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", "")

	// في الإنتاج، JWT_SECRET مطلوب ويجب أن يكون قوي
	if env == "production" {
		if jwtSecret == "" {
			log.Fatal("❌ JWT_SECRET مطلوب في الإنتاج")
		}
		if len(jwtSecret) < 32 {
			log.Fatal("❌ JWT_SECRET يجب أن يكون 32 حرف على الأقل")
		}
	} else if jwtSecret == "" {
		// في التطوير، نولد secret عشوائي إذا لم يتم تحديده
		jwtSecret = generateRandomSecret()
		log.Printf("⚠️  تم توليد JWT_SECRET عشوائي للتطوير (لا تستخدمه في الإنتاج)")
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/ohabits?sslmode=disable"),
		JWTSecret:   jwtSecret,
		Port:        getEnv("PORT", "8080"),
		Env:         env,
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		AIModel:     getEnv("AI_MODEL", "iKhalid/ALLaM:7b"),
	}
}

// generateRandomSecret يولد secret عشوائي قوي
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
