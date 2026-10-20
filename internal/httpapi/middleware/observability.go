package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"automarket/internal/httpapi"
	"github.com/gin-gonic/gin"
)

// RequestLogger records one structured line per request. It never logs bodies,
// cookies, tokens or query strings.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		attributes := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", c.GetHeader("X-Vercel-Id"),
		}
		if user, ok := httpapi.CurrentUser(c); ok {
			attributes = append(attributes, "user_id", user.ID, "role", user.Role)
		}
		logger.InfoContext(c.Request.Context(), "http_request", attributes...)
	}
}

// Recovery converts panics into a logged 500 response.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.ErrorContext(c.Request.Context(), "panic_recovered",
			"panic", fmt.Sprint(recovered),
			"stack", string(debug.Stack()),
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// SecurityHeaders applies conservative defaults for a server-rendered site.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		header.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; img-src 'self' data:; style-src 'self'; script-src 'self'; object-src 'none'")
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			header.Set("Cache-Control", "public, max-age=3600")
		} else {
			header.Set("Cache-Control", "no-store")
		}
		c.Next()
	}
}
