package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/service"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth     httpapi.AuthService
	admin    httpapi.AdminService
	renderer httpapi.Renderer
	logger   *slog.Logger
	cookie   httpapi.CookieConfig
}

func NewAuthHandler(auth httpapi.AuthService, renderer httpapi.Renderer, logger *slog.Logger, cookie httpapi.CookieConfig) *AuthHandler {
	return &AuthHandler{auth: auth, renderer: renderer, logger: logger, cookie: cookie}
}

func (h *AuthHandler) SetAdminService(admin httpapi.AdminService) { h.admin = admin }

type registerPayload struct {
	Name     string `form:"name" json:"name" binding:"required"`
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Next     string `form:"next" json:"next"`
}

func (h *AuthHandler) RegisterSeller(c *gin.Context) {
	var payload registerPayload
	if err := c.ShouldBind(&payload); err != nil {
		page := views.AuthPage{
			BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Registro", "register"),
			Next:     safeNext(payload.Next),
		}
		page.ErrorMessage = "Completa todos los campos."
		h.renderAuthPage(c, http.StatusBadRequest, page)
		return
	}

	_, err := h.auth.RegisterSeller(c.Request.Context(), service.RegisterInput{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		presentation := presentRegistrationError(err)
		if presentation.Internal {
			h.logger.ErrorContext(c.Request.Context(), "register_failed", "error", err)
		}
		page := views.AuthPage{
			BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Registro", "register"),
			Next:     safeNext(payload.Next),
		}
		page.ErrorMessage = presentation.Message
		h.renderAuthPage(c, presentation.Status, page)
		return
	}

	destination := "/login?registered=1"
	if next := safeNext(payload.Next); next != "" {
		destination += "&next=" + url.QueryEscape(next)
	}
	c.Redirect(http.StatusFound, destination)
}

type loginPayload struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Next     string `form:"next" json:"next"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var payload loginPayload
	if err := c.ShouldBind(&payload); err != nil {
		page := views.AuthPage{
			BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Iniciar sesión", "login"),
			Next:     safeNext(payload.Next),
		}
		page.ErrorMessage = "Email y contraseña requeridos."
		h.renderAuthPage(c, http.StatusBadRequest, page)
		return
	}

	result, err := h.auth.Login(c.Request.Context(), payload.Email, payload.Password)
	if err != nil {
		presentation := presentLoginError(err)
		if presentation.Internal {
			h.logger.ErrorContext(c.Request.Context(), "login_failed", "error", err)
		}
		page := views.AuthPage{
			BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Iniciar sesión", "login"),
			Next:     safeNext(payload.Next),
		}
		page.ErrorMessage = presentation.Message
		h.renderAuthPage(c, presentation.Status, page)
		return
	}

	h.setSessionCookie(c, result.Token)
	destination := safeNext(payload.Next)
	if destination == "" {
		destination = "/dashboard"
	}
	c.Redirect(http.StatusFound, destination)
}

// Logout is the POST handler that actually destroys the session.
func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(h.cookie.Name)
	if err == nil && token != "" {
		if err := h.auth.Logout(c.Request.Context(), token); err != nil {
			h.logger.ErrorContext(c.Request.Context(), "logout_failed", "error", err)
		}
	}
	h.clearSessionCookie(c)
	c.Redirect(http.StatusFound, "/login")
}

// LogoutPage refuses to close a session on GET so that link prefetching cannot
// log users out.
func (h *AuthHandler) LogoutPage(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) LoginPage(c *gin.Context) {
	page := views.AuthPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Iniciar sesión", "login"),
		Next:     safeNext(c.Query("next")),
	}
	if c.Query("registered") == "1" {
		page.SuccessMessage = "Cuenta creada correctamente. Ahora puedes iniciar sesión."
	}
	h.renderAuthPage(c, http.StatusOK, page)
}

func (h *AuthHandler) RegisterPage(c *gin.Context) {
	h.renderAuthPage(c, http.StatusOK, views.AuthPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Registro", "register"),
		Next:     safeNext(c.Query("next")),
	})
}

func (h *AuthHandler) Dashboard(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "dashboard.html",
		views.NewDashboardPage(httpapi.CSRFToken(c), &user))
}

type profileResponse struct {
	User views.UserView `json:"user"`
}

func (h *AuthHandler) ShowProfile(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}
	c.JSON(http.StatusOK, profileResponse{User: views.NewUserView(user)})
}

type protectionResponse struct {
	OK    bool        `json:"ok"`
	Role  domain.Role `json:"role"`
	Email string      `json:"email"`
}

func (h *AuthHandler) ProtectedCheck(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}
	c.JSON(http.StatusOK, protectionResponse{OK: true, Role: user.Role, Email: user.Email})
}

type messageResponse struct {
	Message string      `json:"message"`
	Role    domain.Role `json:"role,omitempty"`
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "lista de usuarios", Role: user.Role})
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, token string) {
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(h.cookie.Name, token, int(h.cookie.TTL.Seconds()), "/", "", h.cookie.Secure, true)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(h.cookie.Name, "", -1, "/", "", h.cookie.Secure, true)
}

func (h *AuthHandler) renderAuthPage(c *gin.Context, status int, page views.AuthPage) {
	renderHTML(c, h.renderer, h.logger, status, "auth.html", page)
}

func currentUserPtr(c *gin.Context) *domain.User {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		return nil
	}
	return &user
}

// safeNext accepts only local absolute paths to prevent open redirects.
func safeNext(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.Contains(next, "\\") {
		return ""
	}
	parsed, err := url.Parse(next)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Scheme != "" {
		return ""
	}
	return next
}
