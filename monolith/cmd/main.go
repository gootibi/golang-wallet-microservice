package main

import (
	"log"

	"github.com/gootibi/golang-wallet-microservice/monolith/internal/config"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/database"
)

func main() {
	log.Println("Starting Monolith Wallet Application...")

	// 1. Load configuration
	cfg := config.LoadConfog()

	// 2. Connect to database with retry
	db, err := database.ConnectWithRetry(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Critical Error: could not connect to database after retries: %v", err)
	}
	defer db.Close()

	log.Println("Application successfully initialized...")
}
