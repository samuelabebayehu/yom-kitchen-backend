package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"
)

// GetStatsAdmin retrieves dashboard statistics for the admin panel.
func GetStatsAdmin(c *gin.Context) {
	db := middlewares.GetDBFromContext(c)
	if db == nil {
		c.String(http.StatusInternalServerError, "Database connection not available")
		return
	}

	// --- Fetch Actual Statistics from Database ---
	var menuCount int64
	db.Model(&models.MenuItem{}).Count(&menuCount)
	totalMenus := int(menuCount)

	var orderCount int64
	db.Model(&models.Order{}).Count(&orderCount)
	totalOrders := int(orderCount)

	var clientCount int64
	db.Model(&models.Client{}).Count(&clientCount)
	totalClients := int(clientCount)

	var todayRevenue float64
	today := time.Now().Format("2006-01-02")
	resultRevenue := db.Model(&models.Order{}).
		Where("DATE(created_at) = ?", today).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&todayRevenue)
	if resultRevenue.Error != nil {
		todayRevenue = 0
		log.Printf("Error calculating today's revenue: %v", resultRevenue.Error)
	}
	revenueToday := todayRevenue

	var pendingOrderCount int64
	db.Model(&models.Order{}).
		Where("status = ?", "pending").
		Count(&pendingOrderCount)
	pendingOrders := int(pendingOrderCount)

	// --- Get Order Counts Grouped by Status ---
	var ordersByStatus []struct { // Anonymous struct to hold results
		Status string
		Count  int
	}
	resultStatus := db.Model(&models.Order{}).
		Select("status, COUNT(*) as count"). // Select status and count
		Group("status").                     // Group by status
		Scan(&ordersByStatus)                // Scan results into the struct slice

	if resultStatus.Error != nil {
		log.Printf("Error fetching orders by status: %v", resultStatus.Error)
		ordersByStatus = []struct { // Initialize to empty slice in case of error
			Status string
			Count  int
		}{} // Or handle error differently as needed
	}

	// --- Prepare Statistics Data for JSON Response ---
	stats := gin.H{
		"totalMenus":     totalMenus,
		"totalOrders":    totalOrders,
		"totalClients":   totalClients,
		"revenueToday":   revenueToday,
		"pendingOrders":  pendingOrders,
		"ordersByStatus": ordersByStatus, // Include ordersByStatus in the response
	}

	c.JSON(http.StatusOK, stats)
}
