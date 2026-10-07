package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/config"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/database"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/middleware"
	userHandler "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/handler"
	userRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
	userService "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/service"
)

func main() {
	// Initialize the log
	logger.InnitLogger()
	logger.Log.Info("Starting Monolith Wallet Application...")

	// 1. Load configuration
	cfg := config.LoadConfog()

	// 2. Connect to database with retry
	db, err := database.ConnectWithRetry(cfg.DBDSN)
	if err != nil {
		logger.Log.Error("Critical Error: Could not connect to database after retries", "error", err)
	}
	defer db.Close()

	// 3. Initial layer
	uRepo := userRepository.NewMySQLUserRepository(db)
	uSvc := userService.NewuserService(uRepo)
	uHandler := userHandler.NewUserHandler(uSvc)

	// 4. Setup gin router
	r := gin.Default()

	// Force log's color
	gin.ForceConsoleColor()

	// Attach the logger middleware
	r.Use(gin.Logger())

	// Register global error Handling middleware
	r.Use(middleware.ErrorHandler())

	// Routes grouping
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/users/register", uHandler.Register)
		v1.POST("/users/login", uHandler.Login)
		v1.POST("/users", uHandler.Register)
		v1.GET("/users/:id", uHandler.GetProfile)
		v1.PUT("/users/:id", uHandler.UpdateProfile)

		// Protected routes, only can be accessible if have valid JWT token
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/users/me", uHandler.GetProfileMe)

		}
	}

	// Start server
	logger.Log.Info("Server running on port 8080...")

	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Server failed to run", "error", err)
	}
}
