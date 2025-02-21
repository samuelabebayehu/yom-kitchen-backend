package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateOrderAdmin(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var orderRequest struct {
		ClientID   int `json:"client_id" binding:"required"`
		OrderItems []struct {
			MenuItemID int `json:"menu_item_id" binding:"required"`
			Quantity   int `json:"quantity" binding:"required,min=1"`
		} `json:"order_items" binding:"required,min=1,dive"`
		Notes string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body: "})
		return
	}

	var orderItemsForDB []models.OrderItem
	totalAmount := 0.0

	if err := db.Transaction(func(tx *gorm.DB) error {
		var client models.Client
		if err := tx.First(&client, orderRequest.ClientID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Client ID"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error checking Client: "})
			}
			return err
		}

		for _, itemRequest := range orderRequest.OrderItems {
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, itemRequest.MenuItemID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid MenuItem ID: "})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error checking MenuItem: "})
				}
				return err
			}

			if !menuItem.Available {
				c.JSON(http.StatusBadRequest, gin.H{"message": "MenuItem not available: "})
				return nil
			}

			orderItem := models.OrderItem{
				MenuItemID: itemRequest.MenuItemID,
				ItemName:   menuItem.Name,
				ItemPrice:  menuItem.Price,
				Quantity:   itemRequest.Quantity,
				Subtotal:   menuItem.Price * float64(itemRequest.Quantity),
			}
			orderItemsForDB = append(orderItemsForDB, orderItem)
			totalAmount += orderItem.Subtotal
		}

		order := models.Order{
			ClientID:    orderRequest.ClientID,
			OrderDate:   time.Now(),
			OrderItems:  orderItemsForDB,
			TotalAmount: totalAmount,
			Status:      "Pending",
			Notes:       orderRequest.Notes,
		}
		if createResult := tx.Create(&order); createResult.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create order: "})
			return createResult.Error
		}
		return nil
	}); err != nil {
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully"})
}

func GetOrderAdmin(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid order ID format"})
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var order models.Order
	result := db.Preload("Client").Preload("OrderItems").First(&order, orderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error fetching order: "})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}

func GetAllOrdersAdmin(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var orders []models.Order
	result := db.Preload("Client").Preload("OrderItems").Find(&orders)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error fetching orders: "})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func DeleteOrderAdmin(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid order ID format"})
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var order models.Order
	result := db.First(&order, orderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error checking order: "})
		}
		return
	}

	deleteResult := db.Delete(&order)
	if deleteResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete order: "})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully", "order_id": orderID})
}

func UpdateOrderStatusAdmin(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid order ID format"})
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var order models.Order
	result := db.First(&order, orderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error fetching order: "})
		}
		return
	}

	type OrderStatusUpdateRequest struct {
		Status string `json:"status" binding:"required"`
	}
	var updateRequest OrderStatusUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body: "})
		return
	}

	allowedStatuses := []string{"Pending", "Accepted", "Cancelled", "Ready", "Delivered"}
	isValidStatus := false
	for _, status := range allowedStatuses {
		if updateRequest.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid order status. Allowed statuses: "})
		return
	}

	result = db.Model(&order).UpdateColumn("status", updateRequest.Status)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update order status: "})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update order status (no rows affected)"})
		return
	}

	var updatedOrder models.Order
	db.Preload("Client").Preload("OrderItems").First(&updatedOrder, orderID)

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully", "order": updatedOrder})
}

func stringSliceToString(slice []string) string {
	result := ""
	for i, s := range slice {
		result += s
		if i < len(slice)-1 {
			result += ", "
		}
	}
	return result
}

func ClientCreateOrderHandler(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var orderRequest struct {
		ClientPassword string `json:"passcode" binding:"required"`
		OrderItems     []struct {
			MenuItemID int `json:"menu_item_id" binding:"required"`
			Quantity   int `json:"quantity" binding:"required,min=1"`
		} `json:"order_items" binding:"required,min=1,dive"`
		Notes string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body: "})
		return
	}

	var client models.Client
	if err := db.Where("passcode=?", orderRequest.ClientPassword).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Request"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error checking Client: "})
		}
		return
	}

	if client.Passcode != orderRequest.ClientPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Client Password"})
		return
	}

	var orderItemsForDB []models.OrderItem
	totalAmount := 0.0
	var resolvedClientId = client.ID
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, itemRequest := range orderRequest.OrderItems {
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, itemRequest.MenuItemID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid MenuItem ID: "})
				return err
			}

			if !menuItem.Available {
				c.JSON(http.StatusBadRequest, gin.H{"message": "MenuItem not available: "})
				return nil
			}

			orderItem := models.OrderItem{
				MenuItemID: itemRequest.MenuItemID,
				ItemName:   menuItem.Name,
				ItemPrice:  menuItem.Price,
				Quantity:   itemRequest.Quantity,
				Subtotal:   menuItem.Price * float64(itemRequest.Quantity),
			}
			orderItemsForDB = append(orderItemsForDB, orderItem)
			totalAmount += orderItem.Subtotal
		}

		order := models.Order{
			ClientID:    int(resolvedClientId),
			OrderDate:   time.Now(),
			OrderItems:  orderItemsForDB,
			TotalAmount: totalAmount,
			Status:      "Pending",
			Notes:       orderRequest.Notes,
		}
		if createResult := tx.Create(&order); createResult.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create order: "})
			return createResult.Error
		}
		return nil
	}); err != nil {
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully"})
}

func ClientGetOrdersHandler(c *gin.Context) {
	clientPassword := c.Query("client_password")

	if clientPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Client Password is required"})
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection not available"})
		return
	}

	var client models.Client
	if err := db.Where("passcode=?", clientPassword).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Request"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error checking Client: "})
		}
		return
	}

	if client.Passcode != clientPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Client Password"})
		return
	}
	log.Printf("fetching orders for client ID: %d", client.ID)
	var orders []models.Order
	result := db.Preload("Client").Preload("OrderItems").Where("client_id = ?", client.ID).Find(&orders)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error fetching orders: "})
		return
	}

	c.JSON(http.StatusOK, orders)
}
