// Package handler contains Gin HTTP handlers.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers health-check endpoints.
func RegisterHealthRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
