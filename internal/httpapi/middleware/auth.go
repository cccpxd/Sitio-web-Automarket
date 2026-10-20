package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"github.com/gin-gonic/gin"
)

// RequireAuth rejects requests without a valid server-side session.
func RequireAuth(auth httpapi.SessionValidator, cookieName string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticate(c, auth, cookieName, logger)
		if !ok {
			return
		}
		httpapi.SetCurrentUser(c, user)
		c.Next()
	}
}

// RequireRole authenticates the request and then checks the role allow-list.
func RequireRole(auth httpapi.SessionValidator, cookieName string, logger *slog.Logger, allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticate(c, auth, cookieName, logger)
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

func authenticate(c *gin.Context, auth httpapi.SessionValidator, cookieName string, logger *slog.Logger) (domain.User, bool) {
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return domain.User{}, false
	}

	user, err := auth.Authenticate(c.Request.Context(), token)
	switch {
	case errors.Is(err, domain.ErrInvalidSession):
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión inválida"})
		return domain.User{}, false
	case err != nil:
		logger.ErrorContext(c.Request.Context(), "session_lookup_failed", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return domain.User{}, false
	default:
		return user, true
	}
}
