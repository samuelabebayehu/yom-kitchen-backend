package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"
)

func GetAllClientsAdmin(c *gin.Context) {
	var clients []models.Client
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	result := db.Find(&clients)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: "})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func CreateClientAdmin(context *gin.Context) {
	var newClient models.Client
	if err := context.ShouldBindJSON(&newClient); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "error"})
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}
	var existingClient models.Client
	result := db.Where("name = ? OR email = ?", newClient.Name, newClient.Email).First(&existingClient)
	if result.RowsAffected > 0 {
		print(result.RowsAffected)
		context.JSON(http.StatusBadRequest, gin.H{"message": "Client or email already exists"})
		return
	}

	tx := db.Create(&newClient)
	if tx.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: "})
	}
	context.JSON(http.StatusCreated, newClient)

}

func UpdateClient(context *gin.Context) {

	clientId, err := strconv.Atoi(context.Param("id"))
	db := middlewares.GetDBFromContext(context)

	if db == nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "error"})
		return
	}
	if clientId == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Client id is required"})
		return
	}
	var updatedData models.Client
	if err := context.ShouldBindJSON(&updatedData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "error"})
	}
	log.Println(updatedData)
	var client models.Client
	if result := db.First(&client, clientId); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"message": "client not found"})
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: "})
		}
		return
	}

	result := db.Model(&client).Updates(updatedData)
	if result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update client: "})
		return
	}

	if result.RowsAffected == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update client (no rows affected)"})
		return
	}

	var updatedClient models.Client
	db.First(&updatedClient, clientId)

	context.JSON(http.StatusOK, gin.H{"message": "Client updated successfully", "client": updatedClient})

}

func DeleteClientAdmin(context *gin.Context) {
	clientIdStr := context.Param("id")
	clientId, err := strconv.Atoi(clientIdStr)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid client ID format"})
		return
	}

	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var client models.Client
	result := db.First(&client, clientId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"message": "Client not found"})
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: "})
		}
		return
	}

	deleteResult := db.Delete(&client)
	if deleteResult.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete client: "})
		return
	}

	if deleteResult.RowsAffected == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete client (no rows affected)"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully", "client_id": clientId})
}

func UpdateClientStatusAdmin(context *gin.Context) {
	clientIdStr := context.Param("id")
	clientId, err := strconv.Atoi(clientIdStr)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Client ID format: "})
		return
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	type ClientStatus struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	var clientStatus ClientStatus
	if err := context.ShouldBindJSON(&clientStatus); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body: "})
		return
	}

	var client models.Client
	result := db.First(&client, clientId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"message": "Client item not found"})
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to find client: "})
		}
		return
	}

	updateResult := db.Model(&client).UpdateColumn("available", clientStatus.IsActive)
	if updateResult.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update client availability: "})
		return
	}

	if updateResult.RowsAffected == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to client item availability (no rows affected)"})
		return
	}

	var updatedClient models.Client
	db.First(&updatedClient, clientId)

	context.JSON(http.StatusOK, gin.H{"message": "Client availability updated successfully", "client": updatedClient})
}

func GetClientByIdAdmin(context *gin.Context) {
	clientIdStr := context.Param("id")
	clientId, err := strconv.Atoi(clientIdStr)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid client ID format: "})
		return
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}
	var client models.Client
	result := db.First(&client, clientId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"message": "Client not found"})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to find client: "})
		return
	}
	context.JSON(http.StatusOK, client)
}
