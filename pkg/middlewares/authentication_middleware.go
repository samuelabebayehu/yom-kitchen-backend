package middlewares

import (
	"context"
	"errors"
	"net/http"
	"yom-kitchen/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
)

const UserContextKey = "user"

func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		authToken := parts[1]

		username := authToken

		db := GetDBFromContext(c)
		if db == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		var user models.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token - User not found"})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error during authentication"})
			}
			return
		}

		ctx := context.WithValue(c.Request.Context(), UserContextKey, &user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AdminAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Request.Context().Value(UserContextKey).(*models.User)
		if !exists || user == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized - User information missing"})
			return
		}

		if !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden - Admin access required"})
			return
		}

		c.Next()
	}
}

func GetUserFromContext(c *gin.Context) *models.User {
	user, ok := c.Request.Context().Value(UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil
	}
	return user
}
