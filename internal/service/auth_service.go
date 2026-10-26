package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"automarket/internal/domain"
)

// RegisterInput is the application-level input for seller registration.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

// LoginResult contains the only data the HTTP layer needs after a login.
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      domain.User
}

// AuthService contains authentication business rules. It does not know about
// HTTP, cookies, templates or a specific database.
type AuthService struct {
	users      UserStore
	sessions   SessionStore
	hasher     PasswordHasher
	tokens     TokenIssuer
	ids        IDGenerator
	clock      Clock
	sessionTTL time.Duration
	dummyHash  string
}

func NewAuthService(deps AuthServiceDeps) (*AuthService, error) {
	if deps.Users == nil || deps.Sessions == nil || deps.Hasher == nil ||
		deps.Tokens == nil || deps.IDs == nil || deps.Clock == nil {
		return nil, errors.New("auth service: missing dependency")
	}
	if deps.SessionTTL <= 0 {
		return nil, errors.New("auth service: session TTL must be positive")
	}

	return &AuthService{
		users:      deps.Users,
		sessions:   deps.Sessions,
		hasher:     deps.Hasher,
		tokens:     deps.Tokens,
		ids:        deps.IDs,
		clock:      deps.Clock,
		sessionTTL: deps.SessionTTL,
		dummyHash:  deps.DummyHash,
	}, nil
}

// RegisterSeller validates the input and creates a seller account.
func (s *AuthService) RegisterSeller(ctx context.Context, input RegisterInput) (domain.User, error) {
	registration, err := domain.NormalizeRegistration(input.Name, input.Email, input.Password)
	if err != nil {
		return domain.User{}, err
	}

	return s.createUser(ctx, registration, domain.RoleSeller)
}

// EnsureAdmin creates the bootstrap administrator when it does not exist yet.
// It is idempotent and never promotes an existing account.
func (s *AuthService) EnsureAdmin(ctx context.Context, name, email, password string) (bool, error) {
	registration, err := domain.NormalizeRegistration(name, email, password)
	if err != nil {
		return false, err
	}

	existing, err := s.users.FindByEmail(ctx, registration.Email)
	switch {
	case err == nil:
		if existing.Role != domain.RoleAdmin {
			return false, domain.ErrAdminEmailConflict
		}
		return false, nil
	case !errors.Is(err, domain.ErrNotFound):
		return false, fmt.Errorf("find bootstrap admin: %w", err)
	}

	_, err = s.createUser(ctx, registration, domain.RoleAdmin)
	if errors.Is(err, domain.ErrEmailTaken) {
		// Another instance won the bootstrap race. The account already exists.
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Login verifies credentials and creates a server-side session.
func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	normalizedEmail, err := s.normalizeLoginInput(email, password)
	if err != nil {
		return LoginResult{}, err
	}

	user, err := s.findUserForLogin(ctx, normalizedEmail, password)
	if err != nil {
		return LoginResult{}, err
	}

	if err := s.verifyLoginPassword(user, password); err != nil {
		return LoginResult{}, err
	}

	return s.startSession(ctx, user)
}

// Authenticate resolves a raw token to its user.
func (s *AuthService) Authenticate(ctx context.Context, rawToken string) (domain.User, error) {
	if rawToken == "" {
		return domain.User{}, domain.ErrInvalidSession
	}

	user, err := s.sessions.FindUser(ctx, s.tokens.Fingerprint(rawToken), s.clock.Now())
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrInvalidSession
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find session: %w", err)
	}
	return user, nil
}

// sellerLister is implemented by stores that can enumerate seller accounts.
// It is optional so existing persistence fakes stay small.
type sellerLister interface {
	ListSellers(ctx context.Context) ([]domain.User, error)
}

// ListSellers returns all registered seller accounts for administration.
func (s *AuthService) ListSellers(ctx context.Context) ([]domain.User, error) {
	lister, ok := s.users.(sellerLister)
	if !ok {
		return nil, nil
	}
	sellers, err := lister.ListSellers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sellers: %w", err)
	}
	return sellers, nil
}

// Logout invalidates the session associated with the raw token.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	if err := s.sessions.Delete(ctx, s.tokens.Fingerprint(rawToken)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *AuthService) normalizeLoginInput(email, password string) (string, error) {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" || password == "" {
		s.verifyDummyPassword(password)
		return "", domain.ErrInvalidCredentials
	}
	return normalizedEmail, nil
}

func (s *AuthService) findUserForLogin(ctx context.Context, email, password string) (domain.User, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		s.verifyDummyPassword(password)
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user for login: %w", err)
	}
	return user, nil
}

func (s *AuthService) verifyLoginPassword(user domain.User, password string) error {
	if !s.hasher.Verify(password, user.PasswordHash) {
		return domain.ErrInvalidCredentials
	}
	return nil
}

func (s *AuthService) startSession(ctx context.Context, user domain.User) (LoginResult, error) {
	rawToken, err := s.tokens.Issue()
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue session token: %w", err)
	}

	now := s.clock.Now()
	session := domain.Session{
		TokenHash: s.tokens.Fingerprint(rawToken),
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}

	// Expired sessions are unusable already; cleanup is best-effort and must
	// not block a successful login.
	_ = s.sessions.DeleteExpired(ctx, now)

	if err := s.sessions.Create(ctx, session); err != nil {
		return LoginResult{}, fmt.Errorf("create session: %w", err)
	}

	return LoginResult{
		Token:     rawToken,
		ExpiresAt: session.ExpiresAt,
		User:      user,
	}, nil
}

func (s *AuthService) createUser(ctx context.Context, registration domain.Registration, role domain.Role) (domain.User, error) {
	if _, err := s.users.FindByEmail(ctx, registration.Email); err == nil {
		return domain.User{}, domain.ErrEmailTaken
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, fmt.Errorf("check existing user: %w", err)
	}

	passwordHash, err := s.hasher.Hash(registration.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	id, err := s.ids.New()
	if err != nil {
		return domain.User{}, fmt.Errorf("generate user id: %w", err)
	}

	user := domain.User{
		ID:           id,
		Name:         registration.Name,
		Email:        registration.Email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    s.clock.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *AuthService) verifyDummyPassword(password string) {
	if s.dummyHash == "" {
		return
	}
	s.hasher.Verify(password, s.dummyHash)
}
