package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains every environment-dependent value required by the app.
// It is loaded once at the composition root and injected explicitly.
type Config struct {
	Environment       string
	Port              string
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnectTimeout  time.Duration
	SessionTTL        time.Duration
	CookieName        string
	CookieSecure      bool
	AdminName         string
	AdminEmail        string
	AdminPassword     string
	LogLevel          slog.Level
}

func Load() (Config, error) {
	maxOpenConns, err := intFromEnv("DB_MAX_OPEN_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	maxIdleConns, err := intFromEnv("DB_MAX_IDLE_CONNS", 2)
	if err != nil {
		return Config{}, err
	}
	connMaxLifetime, err := durationFromEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	connectTimeout, err := durationFromEnv("DB_CONNECT_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := durationFromEnv("SESSION_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cookieSecure, err := boolFromEnv("COOKIE_SECURE", true)
	if err != nil {
		return Config{}, err
	}
	logLevel, err := parseLogLevel(envOrDefault("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:       defaultEnvironment(),
		Port:              envOrDefault("PORT", "8080"),
		DatabaseURL:       firstNonEmpty("DATABASE_URL", "POSTGRES_URL", "POSTGRES_PRISMA_URL", "POSTGRES_URL_NON_POOLING"),
		DBMaxOpenConns:    maxOpenConns,
		DBMaxIdleConns:    maxIdleConns,
		DBConnMaxLifetime: connMaxLifetime,
		DBConnectTimeout:  connectTimeout,
		SessionTTL:        sessionTTL,
		CookieName:        envOrDefault("COOKIE_NAME", "session_token"),
		CookieSecure:      cookieSecure,
		AdminName:         envOrDefault("ADMIN_NAME", "Administrador"),
		AdminEmail:        strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		AdminPassword:     os.Getenv("ADMIN_PASSWORD"),
		LogLevel:          logLevel,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c Config) validate() error {
	switch c.Environment {
	case "development", "production":
	default:
		return fmt.Errorf("APP_ENV must be development or production, got %q", c.Environment)
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be a number between 1 and 65535, got %q", c.Port)
	}

	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required (or POSTGRES_URL from the Vercel/Neon integration)")
	}
	if c.DBMaxOpenConns < 1 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be at least 1")
	}
	if c.DBMaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS cannot be negative")
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS cannot exceed DB_MAX_OPEN_CONNS")
	}
	if c.DBConnMaxLifetime <= 0 {
		return fmt.Errorf("DB_CONN_MAX_LIFETIME must be positive")
	}
	if c.DBConnectTimeout <= 0 {
		return fmt.Errorf("DB_CONNECT_TIMEOUT must be positive")
	}
	if c.SessionTTL < 5*time.Minute || c.SessionTTL > 30*24*time.Hour {
		return fmt.Errorf("SESSION_TTL must be between 5m and 720h")
	}
	if strings.TrimSpace(c.CookieName) == "" {
		return fmt.Errorf("COOKIE_NAME cannot be empty")
	}
	if (c.AdminEmail == "") != (c.AdminPassword == "") {
		return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}
	if c.AdminEmail != "" && len(c.AdminPassword) < 12 {
		return fmt.Errorf("ADMIN_PASSWORD must be at least 12 characters")
	}
	if c.IsProduction() {
		if !c.CookieSecure {
			return fmt.Errorf("COOKIE_SECURE must be true in production")
		}
		if strings.Contains(strings.ToLower(c.DatabaseURL), "sslmode=disable") {
			return fmt.Errorf("DATABASE_URL must use TLS in production (remove sslmode=disable)")
		}
	}

	return nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func defaultEnvironment() string {
	if value := strings.TrimSpace(os.Getenv("APP_ENV")); value != "" {
		return value
	}
	if os.Getenv("VERCEL") != "" || os.Getenv("VERCEL_ENV") != "" {
		return "production"
	}
	return "development"
}

func firstNonEmpty(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func intFromEnv(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q", name, value)
	}
	return parsed, nil
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration such as 30s or 24h, got %q", name, value)
	}
	return parsed, nil
}

func boolFromEnv(name string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false, got %q", name, value)
	}
	return parsed, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error, got %q", value)
	}
}
