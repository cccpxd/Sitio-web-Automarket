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

// IDGenerator creates unique identifiers for new records.
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

// CarListStore is the read port for catalog and seller listings.
type CarListStore interface {
	Search(ctx context.Context, filters domain.SearchFilters) ([]domain.Car, int, error)
	FindActive(ctx context.Context, id string) (domain.Car, error)
	FindByID(ctx context.Context, id string) (domain.Car, error)
	ListBySeller(ctx context.Context, sellerID string) ([]domain.Car, error)
	ListAll(ctx context.Context) ([]domain.Car, error)
}

// CarWriteStore is the write port for publishing and management.
type CarWriteStore interface {
	CreateWithPhotos(ctx context.Context, car domain.Car, photos []domain.CarPhoto) error
	UpdateWithPhotos(ctx context.Context, car domain.Car, photos []domain.CarPhoto, replacePhotos bool, cutoff time.Time) error
	Withdraw(ctx context.Context, id, sellerID string, at time.Time) error
	FindByClientToken(ctx context.Context, sellerID, token string) (domain.Car, error)
}

// PhotoStore is the persistence port for photographs.
type PhotoStore interface {
	FindMeta(ctx context.Context, id string) (domain.CarPhoto, error)
	FindData(ctx context.Context, id string) (domain.CarPhoto, error)
}

// ConversationStore is the persistence port for messaging.
type ConversationStore interface {
	Start(ctx context.Context, conversation domain.Conversation, first domain.Message) (domain.Conversation, error)
	Find(ctx context.Context, id string) (domain.Conversation, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Conversation, error)
	AddMessage(ctx context.Context, message domain.Message) (domain.Message, error)
	MarkRead(ctx context.Context, conversationID, readerID string, at time.Time) error
}

// ModerationStore is the persistence port for administrative removals.
type ModerationStore interface {
	RemoveWithAudit(ctx context.Context, entry domain.ModerationEntry) error
	List(ctx context.Context) ([]domain.ModerationEntry, error)
}

// RawPhoto is an uploaded file before validation and normalization.
type RawPhoto struct {
	Name string
	Data []byte
}

// PhotoProcessor validates and normalizes uploaded photographs.
type PhotoProcessor interface {
	Process(raw RawPhoto) (domain.CarPhoto, error)
}
