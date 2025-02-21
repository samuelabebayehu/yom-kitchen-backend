package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const uploadDirectory = "./uploads"

func init() {
	if _, err := os.Stat(uploadDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(uploadDirectory, os.ModeDir|0755)
		if err != nil {
			return
		}
	}
}

func GetAllMenusAdmin(c *gin.Context) {
	var menus []models.MenuItem
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	result := db.Find(&menus)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		return
	}
	c.JSON(http.StatusOK, menus)
}

func CreateMenuAdmin(c *gin.Context) {
	var newMenuItem models.MenuItem

	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		c.String(http.StatusBadRequest, "Parse Multipart Form Error: "+err.Error())
		return
	}
	log.Printf("test print")
	if bindErr := c.ShouldBind(&newMenuItem); bindErr != nil {
		c.String(http.StatusBadRequest, "Bind Error: "+bindErr.Error())
		return
	}
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif"}
		fileContentType := file.Header.Get("Content-Type")
		isValidType := false
		for _, allowedType := range allowedTypes {
			if fileContentType == allowedType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			c.String(http.StatusBadRequest, "Invalid file type. Allowed types: jpeg, png, gif")
			return
		}

		timestamp := time.Now().UnixNano()
		filename := fmt.Sprintf("%d-%s", timestamp, file.Filename)
		filePath := filepath.Join(uploadDirectory, filename)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.String(http.StatusInternalServerError, "Failed to save image: "+err.Error())
			return
		}

		newMenuItem.ImageUrl = "/uploads/" + filename
	} else if !errors.Is(err, http.ErrMissingFile) {
		c.String(http.StatusInternalServerError, "File upload error: "+err.Error())
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var existingMenuItem models.MenuItem
	result := db.Where("name = ? AND category = ?", newMenuItem.Name, newMenuItem.Category).First(&existingMenuItem)
	if result.Error == nil {
		c.String(http.StatusBadRequest, "Menu item already exists")
		return
	}

	tx := db.Create(&newMenuItem)
	if tx.Error != nil {
		c.String(http.StatusInternalServerError, "Database error: "+tx.Error.Error())
		return
	}
	c.JSON(http.StatusCreated, newMenuItem)
}

func UpdateMenuAdmin(c *gin.Context) {
	menuId, err := strconv.Atoi(c.Param("id"))
	db := middlewares.GetDBFromContext(c)

	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if menuId == 0 {
		c.String(http.StatusBadRequest, "Menu id is required")
		return
	}

	var updatedData models.MenuItem
	// Bind form data to MenuItem struct, including text fields
	if err := c.ShouldBind(&updatedData); err != nil { // Use Bind to handle form and JSON
		c.String(http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	// Handle image upload (similar to CreateMenuAdmin)
	file, err := c.FormFile("image") // "image" should match the frontend form field name
	if err == nil && file != nil {   // No error means a new file was uploaded
		// Validate file type (optional, but recommended)
		allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
		fileContentType := file.Header.Get("Content-Type")
		isValidType := false
		for _, allowedType := range allowedTypes {
			if fileContentType == allowedType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			c.String(http.StatusBadRequest, "Invalid file type. Allowed types: jpeg, png, gif")
			return
		}

		// Generate unique filename
		timestamp := time.Now().UnixNano()
		filename := fmt.Sprintf("%d-%s", timestamp, file.Filename)
		filePath := filepath.Join(uploadDirectory, filename)

		// Save file to disk
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.String(http.StatusInternalServerError, "Failed to save image: "+err.Error())
			return
		}

		// Set ImageURL in MenuItem model
		updatedData.ImageUrl = "/uploads/" + filename // Update ImageURL
	} else if !errors.Is(err, http.ErrMissingFile) {
		c.String(http.StatusInternalServerError, "File upload error: "+err.Error())
		return
	} // If http.ErrMissingFile, it means no new image was uploaded, which is okay for update

	var menu models.MenuItem
	if result := db.First(&menu, menuId); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "menu not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		}
		return
	}

	// Update only the fields that are provided in updatedData, including ImageURL if a new image was uploaded
	result := db.Model(&menu).Updates(updatedData)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to update menu: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		c.String(http.StatusInternalServerError, "Failed to update menu (no rows affected)")
		return
	}

	var updatedMenu models.MenuItem
	db.First(&updatedMenu, menuId)

	c.JSON(http.StatusOK, gin.H{"message": "Menu updated successfully", "menu": updatedMenu})
}

func DeleteMenuAdmin(c *gin.Context) {
	menuIdStr := c.Param("id")
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid menu ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var menu models.MenuItem
	result := db.First(&menu, menuId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Menu not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		}
		return
	}

	// Delete associated image file (optional, but recommended for cleanup if storing locally)
	if menu.ImageUrl != "" {
		imagePath := filepath.Join(".", menu.ImageUrl) // Assuming ImageURL is relative to root
		err := os.Remove(imagePath)
		if err != nil {
			return
		} // Remove the image file from disk
		// Handle error if deletion fails, maybe log it but don't block menu deletion
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error deleting image file: %s, error: %v\n", imagePath, err) // Log error, but continue menu deletion
		}
	}

	deleteResult := db.Delete(&menu)
	if deleteResult.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to delete menu: "+deleteResult.Error.Error())
		return
	}

	if deleteResult.RowsAffected == 0 {
		c.String(http.StatusInternalServerError, "Failed to delete menu (no rows affected)")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu deleted successfully", "menu_id": menuId})
}

func UpdateMenuItemAvailabilityAdmin(c *gin.Context) {
	menuIdStr := c.Param("id")
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid menu item ID format: "+err.Error())
		return
	}
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var menuItem models.MenuItem
	result := db.First(&menuItem, menuId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Menu item not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to find menu item: "+result.Error.Error())
		return
	}

	updateResult := db.Model(&menuItem).UpdateColumn("available", !menuItem.Available)
	if updateResult.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to update menu item availability: "+updateResult.Error.Error())
		return
	}

	if updateResult.RowsAffected == 0 {
		c.String(http.StatusInternalServerError, "Failed to update menu item availability (no rows affected)")
		return
	}

	var updatedMenuItem models.MenuItem
	db.First(&updatedMenuItem, menuId)

	c.JSON(http.StatusOK, gin.H{"message": "Menu item availability updated successfully", "menu_item": updatedMenuItem})
}

func GetMenuByIdAdmin(c *gin.Context) {
	menuIdStr := c.Param("id")
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid menu ID format: "+err.Error())
		return
	}
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	var menuItem models.MenuItem
	result := db.First(&menuItem, menuId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Menu item not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to find menu item: "+result.Error.Error())
		return
	}
	c.JSON(http.StatusOK, menuItem)
}

func GetActiveMenus(c *gin.Context) {
	var menus []models.MenuItem
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	result := db.Where("available=true").Find(&menus)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		return
	}
	c.JSON(http.StatusOK, menus)
}
