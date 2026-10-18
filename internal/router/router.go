package router

import (
	"automarket/internal/domain"
	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/repository"
	"automarket/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	repo := repository.NewInMemoryRepository()
	authService := service.NewAuthService(repo)
	authHandler := handlers.NewAuthHandler(authService)

	r.GET("/", handlers.HomePage)
	r.GET("/login", authHandler.LoginPage)
	r.POST("/login", authHandler.Login)
	r.GET("/register", authHandler.RegisterPage)
	r.POST("/register", authHandler.RegisterSeller)
	r.GET("/logout", authHandler.Logout)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth(authService))
	protected.GET("/dashboard", authHandler.Dashboard)
	protected.GET("/profile", authHandler.ShowProfile)
	protected.GET("/protected", authHandler.ProtectedCheck)

	sellerRouter := r.Group("/")
	sellerRouter.Use(middleware.RequireRole(authService, domain.RoleSeller, domain.RoleAdmin))
	sellerRouter.GET("/seller-panel", authHandler.SellerPanel)
	sellerRouter.POST("/cars", authHandler.CreateCar)

	adminRouter := r.Group("/")
	adminRouter.Use(middleware.RequireRole(authService, domain.RoleAdmin))
	adminRouter.GET("/admin-panel", authHandler.AdminPanel)
	adminRouter.GET("/users", authHandler.ListUsers)

	return r
}
