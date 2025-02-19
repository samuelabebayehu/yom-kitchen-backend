package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"
)

func Login(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body: " + err.Error()})
		return
	}

	var user models.User
	result := db.Where("username = ?", loginRequest.Username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error during login: " + result.Error.Error()})
			return
		}
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(time.Hour * 1)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "samuelabebayehu",
		Subject:   strconv.Itoa(int(user.ID)),
	}

	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secretKey) == 0 {
		secretKey = []byte("samuelabebayehu")
		println("WARNING: JWT_SECRET_KEY environment variable not set. Using insecure default key!")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate JWT token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": tokenString, "username": loginRequest.Username})
}
