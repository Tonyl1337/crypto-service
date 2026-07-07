package main

import (
	"log"
	"os"

	"github.com/Tonyl1337/crypto-service/internal/app"
	"github.com/Tonyl1337/crypto-service/internal/config"
	"github.com/Tonyl1337/crypto-service/internal/database"
	"github.com/Tonyl1337/crypto-service/internal/logger"
)

func main() {

	cfg, err := config.Load("configs/config.yaml")
if err != nil {
	log.Fatal(err)
}

logger := logger.New()

db, err := database.New(cfg.Database)
if err != nil {
	logger.Error("failed to connect database", "error", err)
	os.Exit(1)
}

application := app.New(cfg, logger, db)

application.Logger.Info("Database connected")
application.Logger.Info("Application started")

application.Logger.Info("Configuration loaded")

application.Logger.Info(
		"HTTP server", 
		"address", 
		cfg.HTTP.Address,
	)

application.Logger.Info(
		"Database configuration",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
	)
}