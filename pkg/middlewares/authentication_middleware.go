package middlewares

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strconv"
	"yom-kitchen/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
)

const UserContextKey = "user"

func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		parts := strings.Split(tokenString, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		tokenString = parts[1]

		secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
		if len(secretKey) == 0 {
			secretKey = []byte("samuelabebayehu")
			println("WARNING: JWT_SECRET_KEY environment variable not set. Using insecure default key!")
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return secretKey, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userIDString, err := claims.GetSubject()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
			return
		}

		userID, err := strconv.Atoi(userIDString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID in token"})
			return
		}

		db := GetDBFromContext(c)
		if db == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		var user models.User
		result := db.First(&user, userID)
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
