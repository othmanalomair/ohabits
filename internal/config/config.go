package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Env         string
}

func Load() *Config {
	// Load .env file if exists
	godotenv.Load()

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/ohabits?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "default-secret-change-in-production"),
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
	}
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
