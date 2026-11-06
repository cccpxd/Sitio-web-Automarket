package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"github.com/gin-gonic/gin"
)

// RequireAuthPage is like RequireAuth but sends browsers to the login page with
// a validated local return path instead of returning a JSON 401.
func RequireAuthPage(auth httpapi.SessionValidator, cookieName string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticatePage(c, auth, cookieName, logger)
		if !ok {
			return
		}
		httpapi.SetCurrentUser(c, user)
		c.Next()
	}
}

// RequireRolePage authenticates with a login redirect and then enforces the
// role allow-list.
func RequireRolePage(auth httpapi.SessionValidator, cookieName string, logger *slog.Logger, allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticatePage(c, auth, cookieName, logger)
		if !ok {
			return
		}
		if !user.Role.IsAllowed(allowedRoles...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no tienes permisos para esta acción"})
			return
		}
		httpapi.SetCurrentUser(c, user)
		c.Next()
	}
}

func authenticatePage(c *gin.Context, auth httpapi.SessionValidator, cookieName string, logger *slog.Logger) (domain.User, bool) {
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		redirectToLogin(c)
		return domain.User{}, false
	}
	user, err := auth.Authenticate(c.Request.Context(), token)
	switch {
	case errors.Is(err, domain.ErrInvalidSession):
		redirectToLogin(c)
		return domain.User{}, false
	case err != nil:
		logger.ErrorContext(c.Request.Context(), "session_lookup_failed", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return domain.User{}, false
	default:
		return user, true
	}
}

func redirectToLogin(c *gin.Context) {
	target := c.Request.URL.RequestURI()
	parsed, err := url.Parse(target)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		target = "/"
	}
	c.Redirect(http.StatusFound, "/login?next="+url.QueryEscape(target))
}
