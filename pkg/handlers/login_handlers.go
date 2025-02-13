package handlers

import (
	"errors"
	"gorm.io/gorm"
	"net/http"

	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var user models.User
	result := db.Where("username = ?", loginRequest.Username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusUnauthorized, "Invalid username or password")
		} else {
			c.String(http.StatusInternalServerError, "Database error during login: "+result.Error.Error())
		}
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token := loginRequest.Username

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}
