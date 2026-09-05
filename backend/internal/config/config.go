package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	FrontendURL  string
	VAPIDPublic  string
	VAPIDPrivate string
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mias_db?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "mias-super-secret-key-2026-banking-grade-zero-trust")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	vapidPub := getEnv("VAPID_PUBLIC_KEY", "")
	vapidPriv := getEnv("VAPID_PRIVATE_KEY", "")

	return &Config{
		Port:         port,
		DatabaseURL:  dbURL,
		JWTSecret:    jwtSecret,
		FrontendURL:  frontendURL,
		VAPIDPublic:  vapidPub,
		VAPIDPrivate: vapidPriv,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
