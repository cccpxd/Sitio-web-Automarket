package main

import (
	"github.com/gin-gonic/gin"
	"automarket-backend/middleware"
	"automarket-backend/repository"
	"automarket-backend/services"
)

func main() {
	// Inicializar dependencias
	repo := repository.NewMemoryRepo()
	vehicleService := services.NewVehicleService(repo)
	// userService := services.NewUserService(repo)

	// Configurar Gin
	router := gin.Default()

	// Interceptar con Gestor de Seguridad AAA
	protected := router.Group("/")
	protected.Use(middleware.AAAMiddleware())
	
	// Rutas que pasan al controlador/servicio
	protected.POST("/vehicles/:id/approve", func(c *gin.Context) {
		// Simulación de captura y pase al servicio de Vehículos
		// id := capturar ID de la URL...
		vehicleService.ApproveVehicle(1) 
		c.JSON(200, gin.H{"message": "Vehículo aprobado y persistido en memoria"})
	})

	router.Run(":8080")
}