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

func GetAllClientsAdmin(c *gin.Context) {
	var clients []models.Client
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	result := db.Find(&clients)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		return
	}
	c.JSON(http.StatusOK, clients)
}

func CreateClientAdmin(context *gin.Context) {
	var newClient models.Client
	if err := context.ShouldBindJSON(&newClient); err != nil {
		context.String(http.StatusBadRequest, err.Error())
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	var existingClient models.Client
	result := db.Where("name = ? AND email = ?", newClient.Name, newClient.Email).First(&existingClient)
	if result != nil {
		context.String(http.StatusBadRequest, "Client item already exists")
		return
	}

	tx := db.Create(&newClient)
	if tx.Error != nil {
		context.String(http.StatusInternalServerError, "Database error: "+tx.Error.Error())
	}
	context.JSON(http.StatusCreated, newClient)

}

func UpdateClient(context *gin.Context) {
	clientId, err := strconv.Atoi(context.Param("id"))
	db := middlewares.GetDBFromContext(context)

	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	if err != nil {
		context.String(http.StatusBadRequest, err.Error())
		return
	}
	if clientId == 0 {
		context.String(http.StatusBadRequest, "Client id is required")
		return
	}
	var updatedData models.Client
	if err := context.ShouldBindJSON(&updatedData); err != nil {
		context.String(http.StatusBadRequest, err.Error())
	}

	var client models.Client
	if result := db.First(&client, clientId); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "client not found")
		} else {
			context.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		}
		return
	}

	result := db.Model(&client).Updates(updatedData)
	if result.Error != nil {
		context.String(http.StatusInternalServerError, "Failed to update client: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		context.String(http.StatusInternalServerError, "Failed to update client (no rows affected)")
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
		context.String(http.StatusBadRequest, "Invalid client ID format")
		return
	}

	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var client models.Client
	result := db.First(&client, clientId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "Client not found")
		} else {
			context.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
		}
		return
	}

	deleteResult := db.Delete(&client)
	if deleteResult.Error != nil {
		context.String(http.StatusInternalServerError, "Failed to delete client: "+deleteResult.Error.Error())
		return
	}

	if deleteResult.RowsAffected == 0 {
		context.String(http.StatusInternalServerError, "Failed to delete client (no rows affected)")
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully", "client_id": clientId})
}

func UpdateClientStatusAdmin(context *gin.Context) {
	clientIdStr := context.Param("id")
	clientId, err := strconv.Atoi(clientIdStr)
	if err != nil {
		context.String(http.StatusBadRequest, "Invalid Client ID format: "+err.Error())
		return
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	type ClientStatus struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	var clientStatus ClientStatus
	if err := context.ShouldBindJSON(&clientStatus); err != nil {
		context.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var client models.Client
	result := db.First(&client, clientId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "Client item not found")
		} else {
			context.String(http.StatusInternalServerError, "Failed to find client: "+result.Error.Error())
		}
		return
	}

	updateResult := db.Model(&client).UpdateColumn("available", clientStatus.IsActive)
	if updateResult.Error != nil {
		context.String(http.StatusInternalServerError, "Failed to update client availability: "+updateResult.Error.Error())
		return
	}

	if updateResult.RowsAffected == 0 {
		context.String(http.StatusInternalServerError, "Failed to client item availability (no rows affected)")
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
		context.String(http.StatusBadRequest, "Invalid client ID format: "+err.Error())
		return
	}
	db := middlewares.GetDBFromContext(context)
	if db == nil {
		context.String(http.StatusInternalServerError, "Database connection not available")
		return
	}
	var client models.Client
	result := db.First(&client, clientId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context.String(http.StatusNotFound, "Client not found")
			return
		}
		context.String(http.StatusInternalServerError, "Failed to find client: "+result.Error.Error())
		return
	}
	context.JSON(http.StatusOK, client)
}
