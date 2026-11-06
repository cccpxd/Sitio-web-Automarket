package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const csrfContextKey = "csrf_token"

// EnsureCSRFToken makes sure the double-submit cookie exists and exposes its
// value to handlers and templates.
func EnsureCSRFToken(c *gin.Context, config CSRFConfig) {
	token, err := c.Cookie(config.CookieName)
	if err != nil || len(token) < 16 {
		token = newCSRFToken()
		c.SetSameSite(config.SameSite)
		maxAge := int(config.TTL.Seconds())
		c.SetCookie(config.CookieName, token, maxAge, "/", "", config.Secure, false)
	}
	c.Set(csrfContextKey, token)
}

// CSRFToken returns the token for the current request.
func CSRFToken(c *gin.Context) string {
	value, exists := c.Get(csrfContextKey)
	if !exists {
		return ""
	}
	token, _ := value.(string)
	return token
}

// ValidateCSRF compares the submitted form field with the cookie in constant
// time. Safe methods are not checked by the middleware.
func ValidateCSRF(c *gin.Context, config CSRFConfig) bool {
	cookie, err := c.Cookie(config.CookieName)
	if err != nil || cookie == "" {
		return false
	}
	submitted := c.PostForm("_csrf")
	if submitted == "" {
		submitted = c.GetHeader("X-CSRF-Token")
	}
	if submitted == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie), []byte(submitted)) == 1
}

func IsSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func newCSRFToken() string {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "csrf-fallback-token"
	}
	return hex.EncodeToString(buffer)
}
