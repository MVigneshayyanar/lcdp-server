package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr            string
	DatabaseURL     string
	CookieName      string
	CookieDomain    string
	CookieSecure    bool
	CookieSameSite  string
	SessionTTL      time.Duration
	TwoFactorAPIKey string
	Env             string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:            getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		CookieName:      getEnv("COOKIE_NAME", "session"),
		CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
		CookieSameSite:  getEnv("COOKIE_SAMESITE", "Lax"),
		SessionTTL:      getDuration("SESSION_TTL", 24*time.Hour),
		TwoFactorAPIKey: getEnv("TWOFACTOR_API_KEY", ""),
		Env:             getEnv("ENV", "local"),
	}

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}

	cfg.CookieSecure = getCookieSecure()

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	if parsed, err := time.ParseDuration(raw); err == nil {
		return parsed
	}
	return fallback
}

func getCookieSecure() bool {
	env := strings.TrimSpace(os.Getenv("ENV"))

	return strings.EqualFold(env, "prod") || strings.EqualFold(env, "production")
}
