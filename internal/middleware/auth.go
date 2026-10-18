package middleware

import (
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/service"
	"github.com/gin-gonic/gin"
)

func RequireAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
			return
		}

		user, ok := authService.UserFromToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión inválida"})
			return
		}

		c.Set("current_user", user)
		c.Next()
	}
}

func RequireRole(authService *service.AuthService, allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
			return
		}

		user, ok := authService.UserFromToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión inválida"})
			return
		}

		allowed := false
		for _, role := range allowedRoles {
			if user.Role == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no tienes permisos para esta acción"})
			return
		}

		c.Set("current_user", user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) domain.User {
	user, _ := c.Get("current_user")
	if u, ok := user.(domain.User); ok {
		return u
	}
	return domain.User{}
}
