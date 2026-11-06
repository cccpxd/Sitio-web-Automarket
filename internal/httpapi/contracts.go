package httpapi

import (
	"context"
	"io"
	"net/http"
	"time"

	"automarket/internal/domain"
	"automarket/internal/service"
)

// Renderer is the output port used by HTML handlers.
type Renderer interface {
	Render(w io.Writer, name string, data any) error
}

// SessionValidator resolves a raw session token to its user.
type SessionValidator interface {
	Authenticate(ctx context.Context, rawToken string) (domain.User, error)
}

// AuthService is the application port consumed by HTTP handlers.
type AuthService interface {
	SessionValidator
	RegisterSeller(ctx context.Context, input service.RegisterInput) (domain.User, error)
	Login(ctx context.Context, email, password string) (service.LoginResult, error)
	Logout(ctx context.Context, rawToken string) error
}

// CatalogService is the public catalog and search port.
type CatalogService interface {
	Search(ctx context.Context, raw domain.SearchQuery) (domain.SearchResult, error)
	GetActive(ctx context.Context, id string) (domain.Car, error)
	Get(ctx context.Context, id string) (domain.Car, error)
	ListAll(ctx context.Context) ([]domain.Car, error)
}

// PublishingService is the seller listing-management port.
type PublishingService interface {
	Create(ctx context.Context, sellerID string, input service.CreateCarInput) (domain.Car, error)
	Update(ctx context.Context, sellerID, carID string, input service.UpdateCarInput) (domain.Car, error)
	Withdraw(ctx context.Context, sellerID, carID string) error
	ListMine(ctx context.Context, sellerID string) ([]domain.Car, error)
	Get(ctx context.Context, sellerID, carID string) (domain.Car, error)
}

// PhotoService streams stored photographs with visibility rules.
type PhotoService interface {
	GetPublic(ctx context.Context, id string, viewer *domain.User) (domain.CarPhoto, error)
}

// MessagingService is the conversation port.
type MessagingService interface {
	Start(ctx context.Context, carID, buyerID, body string) (domain.Conversation, error)
	Get(ctx context.Context, id string, viewer domain.User) (domain.Conversation, error)
	List(ctx context.Context, userID string) ([]domain.Conversation, error)
	Reply(ctx context.Context, id string, viewer domain.User, body string) (domain.Message, error)
	InitialMessage(car domain.Car) string
}

// ModerationService is the administrative removal port.
type ModerationService interface {
	Remove(ctx context.Context, adminID, carID, reason string) error
	List(ctx context.Context) ([]domain.ModerationEntry, error)
}

// AdminService exposes administration queries.
type AdminService interface {
	ListSellers(ctx context.Context) ([]domain.User, error)
}

// CookieConfig groups the HTTP-only presentation of the session.
type CookieConfig struct {
	Name     string
	Secure   bool
	SameSite http.SameSite
	TTL      time.Duration
}

// CSRFConfig configures the double-submit cookie protection.
type CSRFConfig struct {
	CookieName string
	Secure     bool
	SameSite   http.SameSite
	TTL        time.Duration
}
