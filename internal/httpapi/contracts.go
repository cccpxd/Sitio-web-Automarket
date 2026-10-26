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

// CarService is the application port consumed by HTTP handlers.
type CarService interface {
	ListCars(ctx context.Context) ([]domain.Car, error)
	CreateCar(ctx context.Context, sellerID string, input service.CreateCarInput) (domain.Car, error)
}

// CookieConfig groups the HTTP-only presentation of the session.
type CookieConfig struct {
	Name     string
	Secure   bool
	SameSite http.SameSite
	TTL      time.Duration
}
