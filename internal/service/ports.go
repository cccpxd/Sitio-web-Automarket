package service

import (
	"context"
	"time"

	"automarket/internal/domain"
)

// UserStore is the persistence port required by the authentication service.
type UserStore interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

// SessionStore is the persistence port for server-side sessions.
type SessionStore interface {
	Create(ctx context.Context, session domain.Session) error
	FindUser(ctx context.Context, tokenHash string, now time.Time) (domain.User, error)
	Delete(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

// PasswordHasher hides the password-hashing algorithm from the business layer.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) bool
}

// TokenIssuer creates opaque session tokens and deterministic fingerprints.
type TokenIssuer interface {
	Issue() (string, error)
	Fingerprint(rawToken string) string
}

// IDGenerator creates unique identifiers for new users.
type IDGenerator interface {
	New() (string, error)
}

// Clock makes time-dependent business rules testable.
type Clock interface {
	Now() time.Time
}

// AuthServiceDeps groups the explicit dependencies of AuthService.
type AuthServiceDeps struct {
	Users      UserStore
	Sessions   SessionStore
	Hasher     PasswordHasher
	Tokens     TokenIssuer
	IDs        IDGenerator
	Clock      Clock
	SessionTTL time.Duration
	DummyHash  string
}
