// Package middleware 鉴权中间件。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func OptionalBearer(apiToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiToken == "" {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == apiToken {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
	}
}

func AdminToken(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Admin-Token") != adminToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid admin token", "data": nil})
			return
		}
		c.Next()
	}
}
