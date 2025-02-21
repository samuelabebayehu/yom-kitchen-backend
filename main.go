package main

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"time"
	connection "yom-kitchen/pkg/db"
	"yom-kitchen/pkg/handlers"
	"yom-kitchen/pkg/middlewares"
	"yom-kitchen/pkg/models"
)

const uploadDirectory = "./uploads"

func main() {
	router := gin.Default()
	db, err := connection.InitializeDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	router.Use(middlewares.DatabaseMiddleware(db))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "*"
		},
		MaxAge: 12 * time.Hour,
	}))
	err = db.AutoMigrate(&models.User{}, &models.MenuItem{}, &models.Client{}, &models.Order{}, &models.OrderItem{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
		return
	}
	log.Println("Database migration completed.")

	err = setupAdminUser(db) // Create admin user if not present
	if err != nil {
		log.Fatalf("Error setting up admin user: %v", err)
		return
	}
	log.Println("Admin user setup completed (if needed).")

	router.Static("/uploads", uploadDirectory)
	adminGroup := router.Group("/admin")
	adminGroup.Use(middlewares.AuthenticationMiddleware())
	adminGroup.Use(middlewares.AdminAuthorizationMiddleware())
	{
		stats := adminGroup.Group("/stats")
		{
			stats.GET("", handlers.GetStatsAdmin)
		}
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
		clientRoutes.GET("/menus", handlers.GetActiveMenus)
	}
	router.POST("/login", handlers.Login)
	err = router.Run(":8080")
	if err != nil {
		return
	}

}

func setupAdminUser(db *gorm.DB) error {
	var existingAdminUser models.User
	result := db.Where("is_admin = ?", true).First(&existingAdminUser)
	if result.Error == nil { // Admin user already exists
		log.Println("Admin user already exists, skipping creation.")
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) { // Actual database error
		return result.Error
	}

	var userRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		IsAdmin  bool   `json:"is_admin"`
	}
	userRequest.Username = "admin"
	userRequest.IsAdmin = true
	userRequest.Password = "admin"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error creating User")
		return nil
	}

	var existingUser models.User
	result = db.Unscoped().Where("username = ?", userRequest.Username).First(&existingUser)
	if result.Error == nil {
		log.Printf("Error creating User")
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Error creating User")
		return nil
	}

	newUser := models.User{
		Username:     userRequest.Username,
		PasswordHash: string(hashedPassword),
		IsAdmin:      userRequest.IsAdmin,
	}

	createResult := db.Create(&newUser)
	if createResult.Error != nil {
		log.Printf("Error creating User")
		return nil
	}
	log.Printf("Admin user '%s' created successfully (ID: %d).", newUser.Username, newUser.ID)
	return nil
}
