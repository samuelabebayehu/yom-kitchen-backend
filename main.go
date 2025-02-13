package main

import (
	"github.com/gin-gonic/gin"
	"log"
	connection "yom-kitchen/pkg/db"
	"yom-kitchen/pkg/handlers"
	"yom-kitchen/pkg/middlewares"
)

func main() {
	router := gin.Default()
	db, err := connection.InitializeDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	router.Use(middlewares.DatabaseMiddleware(db))

	adminGroup := router.Group("/admin")
	adminGroup.Use(middlewares.AuthenticationMiddleware())
	adminGroup.Use(middlewares.AdminAuthorizationMiddleware())
	{
		users := adminGroup.Group("/users")
		{
			users.POST("", handlers.CreateUserAdmin)
			users.GET("/:id", handlers.GetUserAdmin)
			users.GET("", handlers.GetAllUsersAdmin)
			users.PUT("/:id", handlers.UpdateUserAdmin)
			users.DELETE("/:id", handlers.DeleteUserAdmin)
		}

		menus := adminGroup.Group("/menus")
		{
			menus.POST("", handlers.CreateMenuAdmin)
			menus.GET("", handlers.GetAllMenusAdmin)
			menus.GET("/:id", handlers.GetMenuByIdAdmin)
			menus.PUT("/:id", handlers.UpdateMenuAdmin)
			menus.DELETE("/:id", handlers.DeleteMenuAdmin)
			menus.PATCH("/:id", handlers.UpdateMenuItemAvailabilityAdmin)
		}

		clients := adminGroup.Group("/clients")
		{
			clients.POST("", handlers.CreateClientAdmin)
			clients.GET("", handlers.GetAllClientsAdmin)
			clients.GET("/:id", handlers.GetClientByIdAdmin)
			clients.PUT("/:id", handlers.UpdateClient)
			clients.DELETE("/:id", handlers.DeleteClientAdmin)
			clients.PATCH("/:id", handlers.UpdateClientStatusAdmin)
		}

		orders := adminGroup.Group("/orders")
		{
			orders.POST("", handlers.CreateOrderAdmin)
			orders.GET("/:id", handlers.GetOrderAdmin)
			orders.GET("", handlers.GetAllOrdersAdmin)
			orders.DELETE("/:id", handlers.DeleteOrderAdmin)
			orders.PUT("/:id/status", handlers.UpdateOrderStatusAdmin)
		}

	}

	clientRoutes := router.Group("/client")
	{
		clientRoutes.POST("/orders", handlers.ClientCreateOrderHandler)
		clientRoutes.GET("/orders", handlers.ClientGetOrdersHandler)
	}
	router.POST("/login", handlers.LoginHandler)
	err = router.Run(":8080")
	if err != nil {
		return
	}

}
