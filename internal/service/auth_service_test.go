package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"automarket/internal/domain"
	"automarket/internal/service"
)

const sessionTTL = time.Hour

type fakeUserStore struct {
	byEmail map[string]domain.User
	byID    map[string]domain.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		byEmail: make(map[string]domain.User),
		byID:    make(map[string]domain.User),
	}
}

func (f *fakeUserStore) Create(_ context.Context, user domain.User) error {
	if _, exists := f.byEmail[user.Email]; exists {
		return domain.ErrEmailTaken
	}
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	return nil
}

func (f *fakeUserStore) FindByEmail(_ context.Context, email string) (domain.User, error) {
	user, exists := f.byEmail[email]
	if !exists {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

type fakeSessionStore struct {
	users  *fakeUserStore
	byHash map[string]domain.Session
}

func newFakeSessionStore(users *fakeUserStore) *fakeSessionStore {
	return &fakeSessionStore{users: users, byHash: make(map[string]domain.Session)}
}

func (f *fakeSessionStore) Create(_ context.Context, session domain.Session) error {
	f.byHash[session.TokenHash] = session
	return nil
}

func (f *fakeSessionStore) FindUser(_ context.Context, tokenHash string, now time.Time) (domain.User, error) {
	session, exists := f.byHash[tokenHash]
	if !exists || !session.ExpiresAt.After(now) {
		return domain.User{}, domain.ErrNotFound
	}
	user, exists := f.users.byID[session.UserID]
	if !exists {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (f *fakeSessionStore) Delete(_ context.Context, tokenHash string) error {
	delete(f.byHash, tokenHash)
	return nil
}

func (f *fakeSessionStore) DeleteExpired(_ context.Context, now time.Time) error {
	for hash, session := range f.byHash {
		if !session.ExpiresAt.After(now) {
			delete(f.byHash, hash)
		}
	}
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (fakeHasher) Verify(password, encodedHash string) bool {
	return encodedHash == "hash:"+password
}

type fakeTokenIssuer struct {
	next int
}

func (f *fakeTokenIssuer) Issue() (string, error) {
	f.next++
	return fmt.Sprintf("raw-token-%d", f.next), nil
}

func (f *fakeTokenIssuer) Fingerprint(rawToken string) string {
	return "fingerprint:" + rawToken
}

type fakeIDGenerator struct {
	next int
}

func (f *fakeIDGenerator) New() (string, error) {
	f.next++
	return fmt.Sprintf("user-%d", f.next), nil
}

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	return f.now
}

func newTestService(t *testing.T) (*service.AuthService, *fakeUserStore, *fakeSessionStore, *fakeClock) {
	t.Helper()

	users := newFakeUserStore()
	sessions := newFakeSessionStore(users)
	clock := &fakeClock{now: time.Date(2026, time.October, 19, 12, 0, 0, 0, time.UTC)}

	authService, err := service.NewAuthService(service.AuthServiceDeps{
		Users:      users,
		Sessions:   sessions,
		Hasher:     fakeHasher{},
		Tokens:     &fakeTokenIssuer{},
		IDs:        &fakeIDGenerator{},
		Clock:      clock,
		SessionTTL: sessionTTL,
		DummyHash:  "hash:dummy",
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return authService, users, sessions, clock
}

func TestRegisterSeller(t *testing.T) {
	authService, _, _, _ := newTestService(t)

	user, err := authService.RegisterSeller(context.Background(), service.RegisterInput{
		Name:     "  Ana   Torres ",
		Email:    " ANA@Example.com ",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if user.Name != "Ana Torres" || user.Email != "ana@example.com" {
		t.Fatalf("unexpected normalized user: %+v", user)
	}
	if user.Role != domain.RoleSeller {
		t.Fatalf("role = %q", user.Role)
	}
	if user.PasswordHash != "hash:secret123" {
		t.Fatalf("password hash = %q", user.PasswordHash)
	}
	if user.ID == "" || user.CreatedAt.IsZero() {
		t.Fatal("id and creation date must be set")
	}
}

func TestRegisterSellerRejectsDuplicateEmail(t *testing.T) {
	authService, _, _, _ := newTestService(t)
	input := service.RegisterInput{Name: "Ana", Email: "ana@example.com", Password: "secret123"}

	if _, err := authService.RegisterSeller(context.Background(), input); err != nil {
		t.Fatalf("first registration: %v", err)
	}
	if _, err := authService.RegisterSeller(context.Background(), input); !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("second registration error = %v", err)
	}
}

func TestRegisterSellerValidatesInput(t *testing.T) {
	authService, _, _, _ := newTestService(t)

	_, err := authService.RegisterSeller(context.Background(), service.RegisterInput{
		Name: "Ana", Email: "not-an-email", Password: "secret123",
	})
	if !errors.Is(err, domain.ErrEmailInvalid) {
		t.Fatalf("error = %v", err)
	}
}

func TestLoginCreatesFingerprintedSession(t *testing.T) {
	authService, _, sessions, clock := newTestService(t)
	registerSeller(t, authService)

	result, err := authService.Login(context.Background(), "ANA@example.com", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Token == "" {
		t.Fatal("token must not be empty")
	}
	if result.ExpiresAt != clock.now.Add(sessionTTL) {
		t.Fatalf("expires at = %v", result.ExpiresAt)
	}

	session, exists := sessions.byHash["fingerprint:"+result.Token]
	if !exists {
		t.Fatal("session was not persisted under the token fingerprint")
	}
	if session.TokenHash == result.Token {
		t.Fatal("raw token must never be persisted")
	}

	user, err := authService.Authenticate(context.Background(), result.Token)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if user.Email != "ana@example.com" {
		t.Fatalf("user email = %q", user.Email)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	authService, _, sessions, _ := newTestService(t)
	registerSeller(t, authService)

	_, err := authService.Login(context.Background(), "ana@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v", err)
	}
	if len(sessions.byHash) != 0 {
		t.Fatal("no session should be created")
	}
}

func TestAuthenticateExpiredSession(t *testing.T) {
	authService, _, _, clock := newTestService(t)
	registerSeller(t, authService)

	result, err := authService.Login(context.Background(), "ana@example.com", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	clock.now = clock.now.Add(sessionTTL + time.Second)
	if _, err := authService.Authenticate(context.Background(), result.Token); !errors.Is(err, domain.ErrInvalidSession) {
		t.Fatalf("error = %v", err)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	authService, _, _, _ := newTestService(t)
	registerSeller(t, authService)

	result, err := authService.Login(context.Background(), "ana@example.com", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if err := authService.Logout(context.Background(), result.Token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := authService.Authenticate(context.Background(), result.Token); !errors.Is(err, domain.ErrInvalidSession) {
		t.Fatalf("error after logout = %v", err)
	}
}

func TestEnsureAdminIsIdempotent(t *testing.T) {
	authService, users, _, _ := newTestService(t)
	ctx := context.Background()

	created, err := authService.EnsureAdmin(ctx, "Admin", "admin@example.com", "super-secret-123")
	if err != nil || !created {
		t.Fatalf("first bootstrap: created=%v err=%v", created, err)
	}
	created, err = authService.EnsureAdmin(ctx, "Admin", "admin@example.com", "super-secret-123")
	if err != nil || created {
		t.Fatalf("second bootstrap: created=%v err=%v", created, err)
	}
	if users.byEmail["admin@example.com"].Role != domain.RoleAdmin {
		t.Fatal("bootstrap user must be admin")
	}
}

func TestEnsureAdminRejectsExistingSeller(t *testing.T) {
	authService, users, _, _ := newTestService(t)
	ctx := context.Background()
	registerSeller(t, authService)

	created, err := authService.EnsureAdmin(ctx, "Ana", "ana@example.com", "super-secret-123")
	if !errors.Is(err, domain.ErrAdminEmailConflict) || created {
		t.Fatalf("bootstrap: created=%v err=%v", created, err)
	}
	if users.byEmail["ana@example.com"].Role != domain.RoleSeller {
		t.Fatal("existing seller must not be promoted")
	}
}

func registerSeller(t *testing.T, authService *service.AuthService) {
	t.Helper()
	_, err := authService.RegisterSeller(context.Background(), service.RegisterInput{
		Name:     "Ana",
		Email:    "ana@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register seller: %v", err)
	}
}
