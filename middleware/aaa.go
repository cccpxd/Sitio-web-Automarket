package middleware

import (
	"fmt"
	"log"
	"time"
	"github.com/gin-gonic/gin"
)
func AAAMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Authentication & Authorization (Simulado)
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "No autorizado"})
			return
		}
		// 2. Accounting (Registrar traza)
		startTime := time.Now()
		c.Next() // Pasa al controlador
		
		latency := time.Since(startTime)
		log.Printf("[Accounting] %s %s | Latencia: %v", c.Request.Method, c.Request.URL.Path, latency)
	}
}