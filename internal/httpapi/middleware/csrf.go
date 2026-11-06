package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"github.com/gin-gonic/gin"
)

// OptionalAuth attaches the current user when a valid session cookie exists,
// but never blocks the request. Public catalog pages use it to adapt the
// navigation without requiring a session.
func OptionalAuth(auth httpapi.SessionValidator, cookieName string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err == nil && token != "" {
			user, err := auth.Authenticate(c.Request.Context(), token)
			switch {
			case err == nil:
				httpapi.SetCurrentUser(c, user)
			case !errors.Is(err, domain.ErrInvalidSession):
				logger.ErrorContext(c.Request.Context(), "optional_session_lookup_failed", "error", err)
			}
		}
		c.Next()
	}
}

// CSRF implements the double-submit cookie pattern. Unsafe methods must send
// the token in the `_csrf` form field or the X-CSRF-Token header.
func CSRF(config httpapi.CSRFConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpapi.EnsureCSRFToken(c, config)
		if httpapi.IsSafeMethod(c.Request.Method) {
			c.Next()
			return
		}
		// Bound the request body before form parsing to avoid unbounded memory
		// use. Individual photo uploads still enforce their own limits.
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<20)
		if !httpapi.ValidateCSRF(c, config) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token CSRF inválido"})
			return
		}
		c.Next()
	}
}
