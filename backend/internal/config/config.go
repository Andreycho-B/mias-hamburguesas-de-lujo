package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	FrontendURL    string
	GoogleClientID string
	VAPIDPublic    string
	VAPIDPrivate   string

	// Variables de producción
	EnvMode    string // "development" | "production"
	DBMaxConns int    // Máximo de conexiones simultáneas al pool PostgreSQL
	CorsOrigin string // Origen permitido explícito para CORS en producción
}

// IsProduction indica si el servidor opera en modo producción.
func (c *Config) IsProduction() bool {
	return c.EnvMode == "production"
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mias_db?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "mias-super-secret-key-2026-banking-grade-zero-trust")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	googleClientID := getEnv("GOOGLE_CLIENT_ID", "98403624230-bt0rlk0cshti1deo9b02egref3g4tjs4.apps.googleusercontent.com")
	vapidPub := getEnv("VAPID_PUBLIC_KEY", "")
	vapidPriv := getEnv("VAPID_PRIVATE_KEY", "")

	envMode := getEnv("ENV_MODE", "development")
	dbMaxConns := getEnvInt("DB_MAX_CONNS", 25)
	corsOrigin := getEnv("CORS_ALLOWED_ORIGIN", "")

	return &Config{
		Port:           port,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		FrontendURL:    frontendURL,
		GoogleClientID: googleClientID,
		VAPIDPublic:    vapidPub,
		VAPIDPrivate:   vapidPriv,
		EnvMode:        envMode,
		DBMaxConns:     dbMaxConns,
		CorsOrigin:     corsOrigin,
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

