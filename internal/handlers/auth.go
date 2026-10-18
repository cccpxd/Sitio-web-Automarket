package handlers

import (
	"net/http"
	"strings"

	"automarket/internal/domain"
	"automarket/internal/middleware"
	"automarket/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterSeller(c *gin.Context) {
	var payload struct {
		Name     string `form:"name" json:"name" binding:"required"`
		Email    string `form:"email" json:"email" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}

	if err := c.ShouldBind(&payload); err != nil {
		c.HTML(http.StatusBadRequest, "auth.html", gin.H{
			"title":        "Registro",
			"errorMessage": "Completa todos los campos.",
			"activePage":   "register",
		})
		return
	}

	user, err := h.authService.RegisterSeller(payload.Name, payload.Email, payload.Password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "auth.html", gin.H{
			"title":        "Registro",
			"errorMessage": err.Error(),
			"activePage":   "register",
		})
		return
	}

	c.HTML(http.StatusCreated, "auth.html", gin.H{
		"title":          "Registro exitoso",
		"successMessage": "Vendedor registrado correctamente. Ahora puedes iniciar sesión.",
		"activePage":     "login",
		"userName":       user.Name,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var payload struct {
		Email    string `form:"email" json:"email" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}

	if err := c.ShouldBind(&payload); err != nil {
		c.HTML(http.StatusBadRequest, "auth.html", gin.H{
			"title":        "Iniciar sesión",
			"errorMessage": "Email y contraseña requeridos.",
			"activePage":   "login",
		})
		return
	}

	token, user, err := h.authService.Login(payload.Email, payload.Password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "auth.html", gin.H{
			"title":        "Iniciar sesión",
			"errorMessage": err.Error(),
			"activePage":   "login",
		})
		return
	}

	c.SetCookie("session_token", token, 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/dashboard")
	_ = user
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil && token != "" {
		_ = h.authService.Logout(token)
	}
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) Dashboard(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.ID == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":    "Dashboard",
		"user":     user,
		"roleText": formatRole(user.Role),
	})
}

func formatRole(role domain.Role) string {
	text := strings.TrimSpace(string(role))
	if text == "" {
		return "Visitante"
	}
	if len(text) == 1 {
		return strings.ToUpper(text)
	}
	return strings.ToUpper(string(text[0])) + text[1:]
}

func (h *AuthHandler) AdminPanel(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.HTML(http.StatusOK, "admin.html", gin.H{
		"title": "Panel de administración",
		"user":  user,
	})
}

func (h *AuthHandler) SellerPanel(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.HTML(http.StatusOK, "seller.html", gin.H{
		"title": "Panel del vendedor",
		"user":  user,
	})
}

func (h *AuthHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "auth.html", gin.H{
		"title":      "Iniciar sesión",
		"activePage": "login",
	})
}

func (h *AuthHandler) RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "auth.html", gin.H{
		"title":      "Registro vendedor",
		"activePage": "register",
	})
}

func (h *AuthHandler) ShowProfile(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) ProtectedCheck(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"ok": true, "role": user.Role, "email": user.Email})
}

func (h *AuthHandler) RequireVisitorRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "visitante autorizado"})
}

func (h *AuthHandler) RequireSellerRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "vendedor autorizado"})
}

func (h *AuthHandler) RequireAdminRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "administrador autorizado"})
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"message": "lista de usuarios", "role": user.Role})
}

func (h *AuthHandler) CreateCar(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"message": "vehículo creado", "role": user.Role})
}

var _ = domain.RoleVisitor
