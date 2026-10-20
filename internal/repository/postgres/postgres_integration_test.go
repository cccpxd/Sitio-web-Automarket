package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"automarket/internal/domain"
	"automarket/internal/repository/postgres"
)

// TestRepositories runs only when TEST_DATABASE_URL points to a disposable
// PostgreSQL database. It validates SQL, migrations and error mapping.
func TestRepositories(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := postgres.Open(ctx, postgres.Options{
		URL:             databaseURL,
		MaxOpenConns:    2,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
		ConnectTimeout:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := postgres.Migrate(ctx, db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := postgres.Migrate(ctx, db); err != nil {
		t.Fatalf("idempotent migration: %v", err)
	}

	users := postgres.NewUserRepository(db)
	sessions := postgres.NewSessionRepository(db)
	now := time.Now().UTC()
	email := fmt.Sprintf("test-%s@example.com", randomHex(t, 8))

	user := domain.User{
		ID:           randomHex(t, 16),
		Name:         "Test User",
		Email:        email,
		PasswordHash: "pbkdf2-sha256$1000$c2FsdHNhbHRzYWx0c2FsdA$ZmFrZWhhc2g",
		Role:         domain.RoleSeller,
		CreatedAt:    now,
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE email = $1`, email)
	})

	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	duplicate := user
	duplicate.ID = randomHex(t, 16)
	if err := users.Create(ctx, duplicate); !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("duplicate error = %v", err)
	}

	found, err := users.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if found.ID != user.ID || found.Role != user.Role {
		t.Fatalf("found user = %+v", found)
	}

	tokenHash := randomHex(t, 32)
	session := domain.Session{
		TokenHash: tokenHash,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}
	if err := sessions.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	sessionUser, err := sessions.FindUser(ctx, tokenHash, now)
	if err != nil {
		t.Fatalf("find user by session: %v", err)
	}
	if sessionUser.ID != user.ID {
		t.Fatalf("session user id = %q", sessionUser.ID)
	}

	if err := sessions.Delete(ctx, tokenHash); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := sessions.FindUser(ctx, tokenHash, now); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted session error = %v", err)
	}

	expiredHash := randomHex(t, 32)
	if err := sessions.Create(ctx, domain.Session{
		TokenHash: expiredHash,
		UserID:    user.ID,
		CreatedAt: now.Add(-2 * time.Hour),
		ExpiresAt: now.Add(-time.Hour),
	}); err != nil {
		t.Fatalf("create expired session: %v", err)
	}
	if err := sessions.DeleteExpired(ctx, now); err != nil {
		t.Fatalf("delete expired sessions: %v", err)
	}
	if _, err := sessions.FindUser(ctx, expiredHash, now); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expired session error = %v", err)
	}
}

func randomHex(t *testing.T, size int) string {
	t.Helper()
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		t.Fatalf("random bytes: %v", err)
	}
	return hex.EncodeToString(buffer)
}
