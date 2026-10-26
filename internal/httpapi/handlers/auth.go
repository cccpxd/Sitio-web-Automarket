package handlers

import (
	"log/slog"
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/service"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth     httpapi.AuthService
	admin    httpapi.AdminService
	cars     httpapi.CarService
	renderer httpapi.Renderer
	logger   *slog.Logger
	cookie   httpapi.CookieConfig
}

func NewAuthHandler(auth httpapi.AuthService, renderer httpapi.Renderer, logger *slog.Logger, cookie httpapi.CookieConfig, cars ...httpapi.CarService) *AuthHandler {
	h := &AuthHandler{
		auth:     auth,
		renderer: renderer,
		logger:   logger,
		cookie:   cookie,
	}
	if len(cars) > 0 {
		h.cars = cars[0]
	}
	return h
}

func (h *AuthHandler) SetCarService(cars httpapi.CarService) {
	h.cars = cars
}

func (h *AuthHandler) SetAdminService(admin httpapi.AdminService) {
	h.admin = admin
}

type registerPayload struct {
	Name     string `form:"name" json:"name" binding:"required"`
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func (h *AuthHandler) RegisterSeller(c *gin.Context) {
	var payload registerPayload
	if err := c.ShouldBind(&payload); err != nil {
		h.renderAuthPage(c, http.StatusBadRequest, views.AuthPage{
			Title:        "Registro",
			ActivePage:   "register",
			ErrorMessage: "Completa todos los campos.",
		})
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
		h.renderAuthPage(c, presentation.Status, views.AuthPage{
			Title:        "Registro",
			ActivePage:   "register",
			ErrorMessage: presentation.Message,
		})
		return
	}

	h.renderAuthPage(c, http.StatusCreated, views.AuthPage{
		Title:          "Registro exitoso",
		ActivePage:     "login",
		SuccessMessage: "Vendedor registrado correctamente. Ahora puedes iniciar sesion.",
	})
}

type loginPayload struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var payload loginPayload
	if err := c.ShouldBind(&payload); err != nil {
		h.renderAuthPage(c, http.StatusBadRequest, views.AuthPage{
			Title:        "Iniciar sesion",
			ActivePage:   "login",
			ErrorMessage: "Email y contrasena requeridos.",
		})
		return
	}

	result, err := h.auth.Login(c.Request.Context(), payload.Email, payload.Password)
	if err != nil {
		presentation := presentLoginError(err)
		if presentation.Internal {
			h.logger.ErrorContext(c.Request.Context(), "login_failed", "error", err)
		}
		h.renderAuthPage(c, presentation.Status, views.AuthPage{
			Title:        "Iniciar sesion",
			ActivePage:   "login",
			ErrorMessage: presentation.Message,
		})
		return
	}

	h.setSessionCookie(c, result.Token)
	c.Redirect(http.StatusFound, "/dashboard")
}

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

func (h *AuthHandler) LoginPage(c *gin.Context) {
	h.renderAuthPage(c, http.StatusOK, views.AuthPage{
		Title:      "Iniciar sesion",
		ActivePage: "login",
	})
}

func (h *AuthHandler) RegisterPage(c *gin.Context) {
	h.renderAuthPage(c, http.StatusOK, views.AuthPage{
		Title:      "Registro vendedor",
		ActivePage: "register",
	})
}

func (h *AuthHandler) Dashboard(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "dashboard.html", views.NewDashboardPage(user))
}

func (h *AuthHandler) AdminPanel(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var sellers []domain.User
	if h.admin != nil {
		list, err := h.admin.ListSellers(c.Request.Context())
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "list_sellers_failed", "error", err)
		} else {
			sellers = list
		}
	}

	var cars []domain.Car
	if h.cars != nil {
		list, err := h.cars.ListCars(c.Request.Context())
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "list_cars_failed", "error", err)
		} else {
			cars = list
		}
	}

	renderHTML(c, h.renderer, h.logger, http.StatusOK, "admin.html", views.NewAdminPage(user, sellers, cars))
}

func (h *AuthHandler) SellerPanel(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	page := views.NewRolePage("Panel del vendedor", user)
	if h.cars != nil {
		cars, err := h.cars.ListCars(c.Request.Context())
		if err == nil {
			page = views.NewRolePageWithCars("Panel del vendedor", user, cars)
		}
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "seller.html", page)
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

type carPayload struct {
	Brand       string `form:"brand" json:"brand" binding:"required"`
	Model       string `form:"model" json:"model" binding:"required"`
	Year        int    `form:"year" json:"year" binding:"required"`
	Description string `form:"description" json:"description"`
	Price       string `form:"price" json:"price" binding:"required"`
}

func (h *AuthHandler) CreateCar(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}

	var payload carPayload
	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos invalidos"})
		return
	}

	if h.cars != nil {
		_, err := h.cars.CreateCar(c.Request.Context(), user.ID, service.CreateCarInput{
			Brand:       payload.Brand,
			Model:       payload.Model,
			Year:        payload.Year,
			Description: payload.Description,
			Price:       payload.Price,
		})
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "create_car_failed", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear el vehiculo"})
			return
		}
	}

	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusCreated, messageResponse{Message: "vehiculo creado", Role: user.Role})
		return
	}
	c.Redirect(http.StatusFound, "/seller-panel")
}

func (h *AuthHandler) DeleteCar(c *gin.Context) {
	if h.cars == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "servicio no disponible"})
		return
	}
	if err := h.cars.DeleteCar(c.Request.Context(), c.Param("id")); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "delete_car_failed", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "publicacion no encontrada"})
		return
	}
	c.Redirect(http.StatusFound, "/admin-panel")
}

func (h *AuthHandler) ReportCar(c *gin.Context) {
	if h.cars == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "servicio no disponible"})
		return
	}
	if err := h.cars.ReportCar(c.Request.Context(), c.Param("id")); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "report_car_failed", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "publicacion no encontrada"})
		return
	}
	c.Redirect(http.StatusFound, "/admin-panel")
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
