package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/auth"
)

type ContextUserKey string

const (
	UserIDKey ContextUserKey = "user_id"
	EmailKey  ContextUserKey = "email"
	AdminKey  ContextUserKey = "admin"
)

func JWTAuth(j *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tok := strings.TrimPrefix(h, "Bearer ")
		claims, err := j.Verify(tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(EmailKey), claims.Email)
		c.Set(string(AdminKey), claims.Admin)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get(string(AdminKey))
		if !exists || v == nil || v.(bool) == false {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}
