package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN string
}

func LoadConfog() *Config {
	// load file .env if there is any
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using enviroment variables")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	return &Config{
		DBDSN: dsn,
	}
}
