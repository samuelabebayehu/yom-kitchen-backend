package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"
)

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

func CreateMenuAdmin(context *gin.Context) {
	var newMenuItem models.MenuItem
	if err := context.ShouldBindJSON(&newMenuItem); err != nil {
		context.String(http.StatusBadRequest, err.Error())
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	var existingMenuItem models.MenuItem
	result := db.Where("name = ? AND category = ?", newMenuItem.Name, newMenuItem.Category).First(&existingMenuItem)
	if result != nil {
		context.String(http.StatusBadRequest, "Menu item already exists")
		return
	}

	tx := db.Create(&newMenuItem)
	if tx.Error != nil {
		context.String(http.StatusInternalServerError, "Database error: "+tx.Error.Error())
	}
	context.JSON(http.StatusCreated, newMenuItem)

}

func UpdateMenuAdmin(context *gin.Context) {
	menuId, err := strconv.Atoi(context.Param("id"))
	db := middlewares.GetDBFromContext(context)

	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	if err != nil {
		context.String(http.StatusBadRequest, err.Error())
		return
	}
	if menuId == 0 {
		context.String(http.StatusBadRequest, "Menu id is required")
		return
	}
	var updatedData models.MenuItem
	if err := context.ShouldBindJSON(&updatedData); err != nil {
		context.String(http.StatusBadRequest, err.Error())
	}

	var menu models.MenuItem
	if result := db.First(&menu, menuId); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "menu not found")
		} else {
			context.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		}
		return
	}

	result := db.Model(&menu).Updates(updatedData)
	if result.Error != nil {
		context.String(http.StatusInternalServerError, "Failed to update menu: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		context.String(http.StatusInternalServerError, "Failed to update menu (no rows affected)")
		return
	}

	var updatedMenu models.MenuItem
	db.First(&updatedMenu, menuId)

	context.JSON(http.StatusOK, gin.H{"message": "Menu updated successfully", "menu": updatedMenu})

}

func DeleteMenuAdmin(context *gin.Context) {
	menuIdStr := context.Param("id")
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		context.String(http.StatusBadRequest, "Invalid menu ID format")
		return
	}

	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var menu models.MenuItem
	result := db.First(&menu, menuId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "Menu not found")
		} else {
			context.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		}
		return
	}

	deleteResult := db.Delete(&menu)
	if deleteResult.Error != nil {
		context.String(http.StatusInternalServerError, "Failed to delete menu: "+deleteResult.Error.Error())
		return
	}

	if deleteResult.RowsAffected == 0 {
		context.String(http.StatusInternalServerError, "Failed to delete menu (no rows affected)")
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Menu deleted successfully", "menu_id": menuId})
}

func UpdateMenuItemAvailabilityAdmin(context *gin.Context) {
	menuIdStr := context.Param("id")
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		context.String(http.StatusBadRequest, "Invalid menu item ID format: "+err.Error())
		return
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	type MenuStatus struct {
		IsAvailable bool `json:"available" binding:"required"`
	}
	var menuStatus MenuStatus
	if err := context.ShouldBindJSON(&menuStatus); err != nil {
		context.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var menuItem models.MenuItem
	result := db.First(&menuItem, menuId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "Menu item not found")
		} else {
			context.String(http.StatusInternalServerError, "Failed to find menu item: "+result.Error.Error())
		}
		return
	}

	updateResult := db.Model(&menuItem).UpdateColumn("available", menuStatus.IsAvailable)
	if updateResult.Error != nil {
		context.String(http.StatusInternalServerError, "Failed to update menu item availability: "+updateResult.Error.Error())
		return
	}

	if updateResult.RowsAffected == 0 {
		context.String(http.StatusInternalServerError, "Failed to update menu item availability (no rows affected)")
		return
	}

	var updatedMenuItem models.MenuItem
	db.First(&updatedMenuItem, menuId)

	context.JSON(http.StatusOK, gin.H{"message": "Menu item availability updated successfully", "menu_item": updatedMenuItem})
}

func GetMenuByIdAdmin(context *gin.Context) {
	menuIdStr := context.Param("id")
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		context.String(http.StatusBadRequest, "Invalid menu ID format: "+err.Error())
		return
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	var menuItem models.MenuItem
	result := db.First(&menuItem, menuId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "Menu item not found")
			return
		}
		context.String(http.StatusInternalServerError, "Failed to find menu item: "+result.Error.Error())
		return
	}
	context.JSON(http.StatusOK, menuItem)
}
