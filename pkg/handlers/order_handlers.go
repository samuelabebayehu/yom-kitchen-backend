package handlers

import (
	"errors"
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
		c.String(http.StatusInternalServerError, "Database connection not available")
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
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var orderItemsForDB []models.OrderItem //
	totalAmount := 0.0

	if err := db.Transaction(func(tx *gorm.DB) error {
		var client models.Client
		if err := tx.First(&client, orderRequest.ClientID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusBadRequest, "Invalid Client ID")
			} else {
				c.String(http.StatusInternalServerError, "Database error checking Client: "+err.Error())
			}
			return err
		}

		for _, itemRequest := range orderRequest.OrderItems {
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, itemRequest.MenuItemID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.String(http.StatusBadRequest, "Invalid MenuItem ID: "+strconv.Itoa(itemRequest.MenuItemID))
				} else {
					c.String(http.StatusInternalServerError, "Database error checking MenuItem: "+err.Error())
				}
				return err
			}

			if !menuItem.Available {
				c.String(http.StatusBadRequest, "MenuItem not available: "+menuItem.Name)
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
			c.String(http.StatusInternalServerError, "Failed to create order: "+createResult.Error.Error())
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
		c.String(http.StatusBadRequest, "Invalid order ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var order models.Order
	result := db.Preload("Client").Preload("OrderItems").First(&order, orderID) // Preload related data
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Order not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error fetching order: "+result.Error.Error())
		}
		return
	}

	c.JSON(http.StatusOK, order)
}

func GetAllOrdersAdmin(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var orders []models.Order
	result := db.Preload("Client").Preload("OrderItems").Find(&orders)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Database error fetching orders: "+result.Error.Error())
		return
	}

	c.JSON(http.StatusOK, orders)
}

func DeleteOrderAdmin(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid order ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var order models.Order
	result := db.First(&order, orderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Order not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error checking order: "+result.Error.Error())
		}
		return
	}

	deleteResult := db.Delete(&order)
	if deleteResult.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to delete order: "+deleteResult.Error.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully", "order_id": orderID})
}

func UpdateOrderStatusAdmin(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid order ID format")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var order models.Order
	result := db.First(&order, orderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Order not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error fetching order: "+result.Error.Error())
		}
		return
	}

	type OrderStatusUpdateRequest struct {
		Status string `json:"status" binding:"required"`
	}
	var updateRequest OrderStatusUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	allowedStatuses := []string{"Pending", "Processing", "Shipped", "Delivered", "Cancelled"}
	isValidStatus := false
	for _, status := range allowedStatuses {
		if updateRequest.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.String(http.StatusBadRequest, "Invalid order status. Allowed statuses: "+stringSliceToString(allowedStatuses))
		return
	}

	result = db.Model(&order).UpdateColumn("status", updateRequest.Status)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Failed to update order status: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		c.String(http.StatusInternalServerError, "Failed to update order status (no rows affected)")
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
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var orderRequest struct {
		ClientID       int    `json:"client_id" binding:"required"`
		ClientPassword string `json:"passcode" binding:"required"` // Client Password in request
		OrderItems     []struct {
			MenuItemID int `json:"menu_item_id" binding:"required"`
			Quantity   int `json:"quantity" binding:"required,min=1"`
		} `json:"order_items" binding:"required,min=1,dive"`
		Notes string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var client models.Client
	if err := db.First(&client, orderRequest.ClientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusBadRequest, "Invalid Client ID")
		} else {
			c.String(http.StatusInternalServerError, "Database error checking Client: "+err.Error())
		}
		return
	}

	if client.Passcode != orderRequest.ClientPassword {
		c.String(http.StatusUnauthorized, "Invalid Client Password")
		return
	}

	var orderItemsForDB []models.OrderItem
	totalAmount := 0.0

	if err := db.Transaction(func(tx *gorm.DB) error {
		var client models.Client
		if err := tx.First(&client, orderRequest.ClientID).Error; err != nil {
			return err
		}

		for _, itemRequest := range orderRequest.OrderItems {
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, itemRequest.MenuItemID).Error; err != nil {
				c.String(http.StatusBadRequest, "Invalid MenuItem ID: "+strconv.Itoa(itemRequest.MenuItemID))
				return err
			}

			if !menuItem.Available {
				c.String(http.StatusBadRequest, "MenuItem not available: "+menuItem.Name)
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
			c.String(http.StatusInternalServerError, "Failed to create order: "+createResult.Error.Error())
			return createResult.Error
		}
		return nil
	}); err != nil {
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully"})
}

func ClientGetOrdersHandler(c *gin.Context) {
	clientIDStr := c.Query("client_id")
	clientPassword := c.Query("client_password")

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Client ID format")
		return
	}

	if clientPassword == "" {
		c.String(http.StatusBadRequest, "Client Password is required")
		return
	}

	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	var client models.Client
	if err := db.First(&client, clientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusBadRequest, "Invalid Client ID")
		} else {
			c.String(http.StatusInternalServerError, "Database error checking Client: "+err.Error())
		}
		return
	}

	if client.Passcode != clientPassword { // Simple password check
		c.String(http.StatusUnauthorized, "Invalid Client Password")
		return
	}

	var orders []models.Order
	result := db.Preload("OrderItems").Where("client_id = ?", clientID).Find(&orders)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Database error fetching orders: "+result.Error.Error())
		return
	}

	c.JSON(http.StatusOK, orders)
}
