package config_test

import (
	"strings"
	"testing"

	"automarket/internal/config"
)

func TestLoadWithValidEnvironment(t *testing.T) {
	clearOptionalEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/automarket")
	t.Setenv("APP_ENV", "production")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/automarket" {
		t.Fatalf("database url = %q", cfg.DatabaseURL)
	}
	if !cfg.CookieSecure {
		t.Fatal("cookies must be secure by default")
	}
	if !cfg.IsProduction() {
		t.Fatal("expected production environment")
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	clearOptionalEnv(t)

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	clearOptionalEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/automarket")
	t.Setenv("SESSION_TTL", "not-a-duration")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "SESSION_TTL") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadDefaultsToProductionOnVercel(t *testing.T) {
	clearOptionalEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/automarket?sslmode=require")
	t.Setenv("VERCEL", "1")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Environment != "production" {
		t.Fatalf("environment = %q", cfg.Environment)
	}
}

func TestLoadRejectsInsecureProductionCookie(t *testing.T) {
	clearOptionalEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/automarket?sslmode=require")
	t.Setenv("APP_ENV", "production")
	t.Setenv("COOKIE_SECURE", "false")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "COOKIE_SECURE") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadRejectsProductionWithoutDatabaseTLS(t *testing.T) {
	clearOptionalEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/automarket?sslmode=disable")
	t.Setenv("APP_ENV", "production")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "TLS") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadRequiresAdminPair(t *testing.T) {
	clearOptionalEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/automarket")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "ADMIN_EMAIL") {
		t.Fatalf("error = %v", err)
	}
}

func clearOptionalEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"APP_ENV", "PORT", "DATABASE_URL", "POSTGRES_URL", "POSTGRES_PRISMA_URL",
		"POSTGRES_URL_NON_POOLING", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME", "DB_CONNECT_TIMEOUT", "SESSION_TTL", "COOKIE_NAME",
		"COOKIE_SECURE", "ADMIN_NAME", "ADMIN_EMAIL", "ADMIN_PASSWORD", "LOG_LEVEL",
		"VERCEL", "VERCEL_ENV",
	} {
		t.Setenv(name, "")
	}
}
