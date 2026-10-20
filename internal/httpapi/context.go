package httpapi

import (
	"automarket/internal/domain"
	"github.com/gin-gonic/gin"
)

// CurrentUserKey is the Gin context key used by the authentication middleware.
const CurrentUserKey = "current_user"

func SetCurrentUser(c *gin.Context, user domain.User) {
	c.Set(CurrentUserKey, user)
}

func CurrentUser(c *gin.Context) (domain.User, bool) {
	value, exists := c.Get(CurrentUserKey)
	if !exists {
		return domain.User{}, false
	}
	user, ok := value.(domain.User)
	return user, ok
}
