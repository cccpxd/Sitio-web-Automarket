package router

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/httpapi/handlers"
	"automarket/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Auth        httpapi.AuthService
	Cars        httpapi.CarService
	Renderer    httpapi.Renderer
	StaticFS    fs.FS
	Logger      *slog.Logger
	Cookie      httpapi.CookieConfig
	HealthCheck func(ctx context.Context) error
}

func New(deps Dependencies) (*gin.Engine, error) {
	if deps.Auth == nil || deps.Renderer == nil || deps.StaticFS == nil || deps.Logger == nil {
		return nil, errors.New("router: missing dependency")
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	engine.Use(
		middleware.Recovery(deps.Logger),
		middleware.RequestLogger(deps.Logger),
		middleware.SecurityHeaders(),
	)
	engine.StaticFS("/static", http.FS(deps.StaticFS))

	pageHandler := handlers.NewPageHandler(deps.Renderer, deps.Logger, deps.HealthCheck, deps.Cars)
	authHandler := handlers.NewAuthHandler(deps.Auth, deps.Renderer, deps.Logger, deps.Cookie, deps.Cars)

	engine.GET("/", pageHandler.Home)
	engine.GET("/healthz", pageHandler.Health)
	engine.GET("/login", authHandler.LoginPage)
	engine.POST("/login", authHandler.Login)
	engine.GET("/register", authHandler.RegisterPage)
	engine.POST("/register", authHandler.RegisterSeller)
	engine.GET("/logout", authHandler.Logout)

	protected := engine.Group("/")
	protected.Use(middleware.RequireAuth(deps.Auth, deps.Cookie.Name, deps.Logger))
	protected.GET("/dashboard", authHandler.Dashboard)
	protected.GET("/profile", authHandler.ShowProfile)
	protected.GET("/protected", authHandler.ProtectedCheck)

	seller := engine.Group("/")
	seller.Use(middleware.RequireRole(deps.Auth, deps.Cookie.Name, deps.Logger, domain.RoleSeller, domain.RoleAdmin))
	seller.GET("/seller-panel", authHandler.SellerPanel)
	seller.POST("/cars", authHandler.CreateCar)

	admin := engine.Group("/")
	admin.Use(middleware.RequireRole(deps.Auth, deps.Cookie.Name, deps.Logger, domain.RoleAdmin))
	admin.GET("/admin-panel", authHandler.AdminPanel)
	admin.GET("/users", authHandler.ListUsers)

	return engine, nil
}
