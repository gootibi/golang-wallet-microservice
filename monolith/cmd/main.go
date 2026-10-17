package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/config"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/database"
	ledgerRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/repository"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/middleware"
	txHandler "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/handler"
	txRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/repository"
	txService "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/service"
	userHandler "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/handler"
	userRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
	userService "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/service"
	walletHandler "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/handler"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	walletService "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/service"
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
	// Repository
	uRepo := userRepository.NewMySQLUserRepository(db)
	wRepo := walletRepository.NewMySQLWalletRepository(db)
	tRepo := txRepository.NewMySQLTransactionRepositor(db)
	lRepo := ledgerRepository.NewMySQLLedgerRepository(db)

	// Service
	uSvc := userService.NewuserService(db, uRepo, wRepo) // Inject db to user service for transaction
	wSvc := walletService.NewWalletService(wRepo)
	tSvc := txService.NewTransactionService(db, tRepo, uRepo, wRepo, lRepo)

	// Handler
	uHandler := userHandler.NewUserHandler(uSvc)
	wHandler := walletHandler.NewWalletHandler(wSvc)
	tHandler := txHandler.NewTransactionHandler(tSvc)

	// 4. Setup gin router
	// r := gin.Default()
	r := gin.New()

	// Force log's color
	gin.ForceConsoleColor()

	// Recovery middleware
	r.Use(gin.Recovery())

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
		// v1.POST("/users", uHandler.Register)
		// v1.GET("/users/:id", uHandler.GetProfile)
		// v1.PUT("/users/:id", uHandler.UpdateProfile)

		// Protected routes, only can be accessible if have valid JWT token
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User
			protected.GET("/users/me", uHandler.GetProfileMe)
			protected.POST("/users/avatar", uHandler.UploadAvatar)
			protected.DELETE("users/me", uHandler.DeleteAccount)

			// Wallet
			protected.GET("/wallets/me", wHandler.GetMyWallet)

			// Transaction
			protected.POST("/transactions/transfer", tHandler.Transfer)
			protected.GET("/transactions/history", tHandler.GetHistory)
		}
	}

	// Start server
	logger.Log.Info("Server running on port 8080...")

	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Server failed to run", "error", err)
	}
}
