package router

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/httpapi/handlers"
	"automarket/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Auth        httpapi.AuthService
	Admin       httpapi.AdminService
	Catalog     httpapi.CatalogService
	Publishing  httpapi.PublishingService
	Photos      httpapi.PhotoService
	Messages    httpapi.MessagingService
	Moderation  httpapi.ModerationService
	Renderer    httpapi.Renderer
	StaticFS    fs.FS
	Logger      *slog.Logger
	Cookie      httpapi.CookieConfig
	CSRF        httpapi.CSRFConfig
	HealthCheck func(ctx context.Context) error
}

func New(deps Dependencies) (*gin.Engine, error) {
	if deps.Auth == nil || deps.Renderer == nil || deps.StaticFS == nil || deps.Logger == nil {
		return nil, errors.New("router: missing dependency")
	}
	if deps.CSRF.CookieName == "" {
		deps.CSRF = httpapi.CSRFConfig{
			CookieName: "csrf_token",
			Secure:     deps.Cookie.Secure,
			SameSite:   http.SameSiteLaxMode,
			TTL:        deps.Cookie.TTL,
		}
	}
	if deps.CSRF.TTL <= 0 {
		deps.CSRF.TTL = 24 * time.Hour
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
		middleware.OptionalAuth(deps.Auth, deps.Cookie.Name, deps.Logger),
		middleware.CSRF(deps.CSRF),
	)
	engine.StaticFS("/static", http.FS(deps.StaticFS))

	pageHandler := handlers.NewPageHandler(deps.Renderer, deps.Logger, deps.HealthCheck, deps.Catalog)
	authHandler := handlers.NewAuthHandler(deps.Auth, deps.Renderer, deps.Logger, deps.Cookie)
	authHandler.SetAdminService(deps.Admin)
	publishingHandler := handlers.NewPublishingHandler(deps.Publishing, deps.Catalog, deps.Renderer, deps.Logger)
	photoHandler := handlers.NewPhotoHandler(deps.Photos, deps.Logger)
	messageHandler := handlers.NewMessageHandler(deps.Messages, deps.Catalog, deps.Renderer, deps.Logger)
	moderationHandler := handlers.NewModerationHandler(deps.Moderation, deps.Catalog, deps.Admin, deps.Renderer, deps.Logger)

	engine.GET("/", pageHandler.Home)
	engine.GET("/cars", pageHandler.Catalog)
	engine.GET("/cars/:id", pageHandler.Detail)
	engine.GET("/cars/:id/contacto", messageHandler.ContactForm)
	engine.GET("/photos/:id", photoHandler.Serve)
	engine.GET("/healthz", pageHandler.Health)
	engine.GET("/login", authHandler.LoginPage)
	engine.POST("/login", authHandler.Login)
	engine.GET("/register", authHandler.RegisterPage)
	engine.POST("/register", authHandler.RegisterSeller)
	engine.POST("/logout", authHandler.Logout)
	engine.GET("/logout", authHandler.LogoutPage)

	protected := engine.Group("/")
	protected.Use(middleware.RequireAuth(deps.Auth, deps.Cookie.Name, deps.Logger))
	protected.GET("/dashboard", authHandler.Dashboard)
	protected.GET("/profile", authHandler.ShowProfile)
	protected.GET("/protected", authHandler.ProtectedCheck)

	// Browser pages use login redirects so that an anonymous visitor can write
	// a message draft and return to it after authenticating.
	pageProtected := engine.Group("/")
	pageProtected.Use(middleware.RequireAuthPage(deps.Auth, deps.Cookie.Name, deps.Logger))
	pageProtected.GET("/mensajes", messageHandler.Inbox)
	pageProtected.GET("/mensajes/:id", messageHandler.Thread)
	pageProtected.POST("/mensajes/:id", messageHandler.Reply)
	pageProtected.POST("/cars/:id/contacto", messageHandler.ContactSubmit)

	seller := engine.Group("/")
	seller.Use(middleware.RequireRolePage(deps.Auth, deps.Cookie.Name, deps.Logger, domain.RoleSeller))
	seller.GET("/seller-panel", func(c *gin.Context) { c.Redirect(http.StatusFound, "/seller/ventas") })
	seller.GET("/seller/ventas", publishingHandler.Sales)
	seller.GET("/seller/publicar", publishingHandler.PublishForm)
	seller.POST("/seller/publicar", publishingHandler.Create)
	seller.GET("/seller/publicaciones/:id/editar", publishingHandler.EditForm)
	seller.POST("/seller/publicaciones/:id/editar", publishingHandler.Update)
	seller.POST("/seller/publicaciones/:id/retirar", publishingHandler.Withdraw)

	admin := engine.Group("/")
	admin.Use(middleware.RequireRole(deps.Auth, deps.Cookie.Name, deps.Logger, domain.RoleAdmin))
	admin.GET("/admin-panel", moderationHandler.Panel)
	admin.GET("/admin/moderacion", moderationHandler.Panel)
	admin.GET("/admin/publicaciones/:id/retirar", moderationHandler.Confirm)
	admin.POST("/admin/publicaciones/:id/retirar", moderationHandler.Remove)
	admin.GET("/admin/conversaciones/:id", messageHandler.Thread)
	admin.GET("/users", authHandler.ListUsers)

	return engine, nil
}
