package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateUserAdmin(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var userRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to hash password: "+err.Error())
		return
	}

	var existingUser models.User
	result := db.Where("username = ?", userRequest.Username).First(&existingUser)
	if result.Error == nil {
		c.String(http.StatusConflict, "Username already exists")
		return
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.String(http.StatusInternalServerError, "Database error checking username: "+result.Error.Error())
		return
	}

	newUser := models.User{
		Username:     userRequest.Username,
		PasswordHash: string(hashedPassword),
		IsAdmin:      userRequest.IsAdmin,
	}

	createResult := db.Create(&newUser)
	if createResult.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to create user: "+createResult.Error.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user_id": newUser.ID, "username": newUser.Username})
}

func GetUserAdmin(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid user ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var user models.User
	result := db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "User not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error fetching user: "+result.Error.Error())
		}
		return
	}

	userResponse := struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Username  string    `json:"username"`
		IsAdmin   bool      `json:"is_admin"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
	}

	c.JSON(http.StatusOK, userResponse)
}

func GetAllUsersAdmin(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var users []models.User
	result := db.Find(&users)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Database error fetching users: "+result.Error.Error())
		return
	}

	var usersResponse []interface{}
	for _, user := range users {
		usersResponse = append(usersResponse, struct {
			ID        uint      `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Username  string    `json:"username"`
			IsAdmin   bool      `json:"is_admin"`
		}{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Username:  user.Username,
			IsAdmin:   user.IsAdmin,
		})
	}

	c.JSON(http.StatusOK, usersResponse)
}

func UpdateUserAdmin(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid user ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var user models.User
	result := db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "User not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error fetching user: "+result.Error.Error())
		}
		return
	}

	var userRequest struct {
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
		IsAdmin  *bool   `json:"is_admin,omitempty"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	updates := make(map[string]interface{})

	if userRequest.Username != nil {
		var existingUser models.User
		usernameCheckResult := db.Where("username = ? AND id != ?", *userRequest.Username, userID).First(&existingUser)
		if usernameCheckResult.Error == nil {
			c.String(http.StatusConflict, "Username already exists")
			return
		}
		if !errors.Is(usernameCheckResult.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusInternalServerError, "Database error checking username: "+usernameCheckResult.Error.Error())
			return
		}
		updates["username"] = *userRequest.Username
	}

	if userRequest.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*userRequest.Password), bcrypt.DefaultCost)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to hash new password: "+err.Error())
			return
		}
		updates["password_hash"] = string(hashedPassword)
	}

	if userRequest.IsAdmin != nil {
		updates["is_admin"] = *userRequest.IsAdmin
	}

	if len(updates) > 0 {
		updateResult := db.Model(&user).Updates(updates)
		if updateResult.Error != nil {
			c.String(http.StatusInternalServerError, "Failed to update user: "+updateResult.Error.Error())
			return
		}
		if updateResult.RowsAffected == 0 {
			c.String(http.StatusOK, "User updated successfully (no changes applied)")
			return
		}
	}

	var updatedUser models.User
	db.First(&updatedUser, userID)

	userResponse := struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Username  string    `json:"username"`
		IsAdmin   bool      `json:"is_admin"`
	}{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Username:  updatedUser.Username,
		IsAdmin:   updatedUser.IsAdmin,
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": userResponse})
}

func DeleteUserAdmin(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid user ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var user models.User
	result := db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "User not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error fetching user: "+result.Error.Error())
		}
		return
	}

	deleteResult := db.Delete(&user)
	if deleteResult.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to delete user: "+deleteResult.Error.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "user_id": userID})
}
