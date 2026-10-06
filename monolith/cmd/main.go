package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/config"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/database"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
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

	// Routes
	r.POST("/api/v1/users", uHandler.Register)
	r.GET("/api/v1/users/:id", uHandler.GetProfile)
	r.PUT("/api/v1/users/:id", uHandler.UpdateProfile)

	// Start server
	logger.Log.Info("Server running on port 8080...")

	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Server failed to run", "error", err)
	}
}
