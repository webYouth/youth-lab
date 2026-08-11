// Package handler contains Gin HTTP handlers.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers health-check endpoints.
func RegisterHealthRoutes(router *gin.Engine) {
	health := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	// Direct path used inside docker network / local.
	router.GET("/health", health)
	// Compatibility path in case nginx does not strip /api prefix.
	router.GET("/api/health", health)
}
