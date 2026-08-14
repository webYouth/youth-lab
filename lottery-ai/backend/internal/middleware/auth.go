package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"youthlab/lottery-ai/internal/auth"
)

func OptionalBearer(apiToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiToken == "" {
			c.Next()
			return
		}
		h := c.GetHeader("Authorization")
		if strings.HasPrefix(h, "Bearer ") && strings.TrimPrefix(h, "Bearer ") == apiToken {
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

func RequireUser(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录", "data": nil})
			return
		}
		claims, err := auth.ParseToken(jwtSecret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录已失效，请重新登录", "data": nil})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
