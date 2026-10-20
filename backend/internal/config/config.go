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
	jwtSecret := getEnv("JWT_SECRET", "mias-dev-secret-key-32-chars-long!")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	googleClientID := getEnv("GOOGLE_CLIENT_ID", "")
	vapidPub := getEnv("VAPID_PUBLIC_KEY", "")
	vapidPriv := getEnv("VAPID_PRIVATE_KEY", "")

	envMode := getEnv("ENV_MODE", "development")
	dbMaxConns := getEnvInt("DB_MAX_CONNS", 25)
	corsOrigin := getEnv("CORS_ALLOWED_ORIGIN", "")

	cfg := &Config{
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

	if cfg.IsProduction() && os.Getenv("JWT_SECRET") == "" {
		panic("JWT_SECRET must be explicitly set when ENV_MODE is production")
	}

	return cfg
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

