package router_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/httpapi/router"
	"automarket/internal/service"
	"automarket/internal/webui"
	"github.com/gin-gonic/gin"
)

type fakeAuth struct {
	users    map[string]domain.User
	sessions map[string]string
}

func newFakeAuth() *fakeAuth {
	return &fakeAuth{
		users: map[string]domain.User{
			"ana@example.com": {
				ID:           "user-1",
				Name:         "Ana Torres",
				Email:        "ana@example.com",
				PasswordHash: "hash:secret123",
				Role:         domain.RoleSeller,
				CreatedAt:    time.Date(2026, time.October, 19, 12, 0, 0, 0, time.UTC),
			},
		},
		sessions: make(map[string]string),
	}
}

func (f *fakeAuth) Authenticate(_ context.Context, rawToken string) (domain.User, error) {
	email, exists := f.sessions[rawToken]
	if !exists {
		return domain.User{}, domain.ErrInvalidSession
	}
	user, exists := f.users[email]
	if !exists {
		return domain.User{}, domain.ErrInvalidSession
	}
	return user, nil
}

func (f *fakeAuth) RegisterSeller(_ context.Context, input service.RegisterInput) (domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, exists := f.users[email]; exists {
		return domain.User{}, domain.ErrEmailTaken
	}
	user := domain.User{
		ID:           "user-2",
		Name:         input.Name,
		Email:        email,
		PasswordHash: "hash:" + input.Password,
		Role:         domain.RoleSeller,
		CreatedAt:    time.Now().UTC(),
	}
	f.users[email] = user
	return user, nil
}

func (f *fakeAuth) Login(_ context.Context, email, password string) (service.LoginResult, error) {
	user, exists := f.users[strings.ToLower(strings.TrimSpace(email))]
	if !exists || user.PasswordHash != "hash:"+password {
		return service.LoginResult{}, domain.ErrInvalidCredentials
	}
	token := "valid-token"
	f.sessions[token] = user.Email
	return service.LoginResult{
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		User:      user,
	}, nil
}

func (f *fakeAuth) Logout(_ context.Context, rawToken string) error {
	delete(f.sessions, rawToken)
	return nil
}

func newTestRouter(t *testing.T, health func(context.Context) error) *gin.Engine {
	t.Helper()

	renderer, err := webui.NewRenderer()
	if err != nil {
		t.Fatalf("new renderer: %v", err)
	}
	staticFS, err := webui.StaticFS()
	if err != nil {
		t.Fatalf("static fs: %v", err)
	}

	engine, err := router.New(router.Dependencies{
		Auth:        newFakeAuth(),
		Renderer:    renderer,
		StaticFS:    staticFS,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		HealthCheck: health,
		Cookie: httpapi.CookieConfig{
			Name:     "session_token",
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			TTL:      time.Hour,
		},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return engine
}

func TestPublicRoutes(t *testing.T) {
	engine := newTestRouter(t, nil)

	response := performRequest(engine, http.MethodGet, "/", nil, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("home status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "AutoMarket") {
		t.Fatal("home page did not render")
	}

	loginPage := performRequest(engine, http.MethodGet, "/login", nil, nil)
	if loginPage.Code != http.StatusOK || !strings.Contains(loginPage.Body.String(), "Iniciar sesión") {
		t.Fatalf("login page status = %d", loginPage.Code)
	}

	staticFile := performRequest(engine, http.MethodGet, "/static/css/style.css", nil, nil)
	if staticFile.Code != http.StatusOK || !strings.Contains(staticFile.Body.String(), "--primary") {
		t.Fatalf("static file status = %d", staticFile.Code)
	}

	health := performRequest(engine, http.MethodGet, "/healthz", nil, nil)
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d", health.Code)
	}
}

func TestProtectedRouteRequiresSession(t *testing.T) {
	engine := newTestRouter(t, nil)

	response := performRequest(engine, http.MethodGet, "/dashboard", nil, nil)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestLoginCreatesCookieAndAllowsDashboard(t *testing.T) {
	engine := newTestRouter(t, nil)

	form := url.Values{"email": {"ana@example.com"}, "password": {"secret123"}}
	login := performRequest(engine, http.MethodPost, "/login", strings.NewReader(form.Encode()), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if login.Code != http.StatusFound {
		t.Fatalf("login status = %d, body=%s", login.Code, login.Body.String())
	}

	cookies := login.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a session cookie")
	}

	dashboard := performRequest(engine, http.MethodGet, "/dashboard", nil, nil, cookies[0])
	if dashboard.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d", dashboard.Code)
	}
	if !strings.Contains(dashboard.Body.String(), "Ana Torres") {
		t.Fatal("dashboard did not render the authenticated user")
	}
}

func TestProfileDoesNotExposePasswordHash(t *testing.T) {
	engine := newTestRouter(t, nil)
	cookie := loginCookie(t, engine)

	response := performRequest(engine, http.MethodGet, "/profile", nil, nil, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("profile status = %d", response.Code)
	}

	body := response.Body.String()
	if strings.Contains(body, "hash:secret123") || strings.Contains(body, "PasswordHash") {
		t.Fatalf("profile leaked password data: %s", body)
	}
	if !strings.Contains(body, "ana@example.com") {
		t.Fatalf("profile response is missing user data: %s", body)
	}
}

func TestSellerCannotAccessAdminPanel(t *testing.T) {
	engine := newTestRouter(t, nil)
	cookie := loginCookie(t, engine)

	response := performRequest(engine, http.MethodGet, "/admin-panel", nil, nil, cookie)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestHealthReportsDependencyFailure(t *testing.T) {
	engine := newTestRouter(t, func(context.Context) error {
		return errors.New("database unavailable")
	})

	response := performRequest(engine, http.MethodGet, "/healthz", nil, nil)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
}

func loginCookie(t *testing.T, engine *gin.Engine) *http.Cookie {
	t.Helper()
	form := url.Values{"email": {"ana@example.com"}, "password": {"secret123"}}
	login := performRequest(engine, http.MethodPost, "/login", strings.NewReader(form.Encode()), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	cookies := login.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("login status = %d, no cookie", login.Code)
	}
	return cookies[0]
}

func performRequest(engine *gin.Engine, method, path string, body io.Reader, headers map[string]string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, body)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}
