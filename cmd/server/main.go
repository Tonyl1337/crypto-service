package main

import (
	"log"

	"github.com/Tony1337/crypto-service/internal/config"
	"github.com/Tony1337/crypto-service/internal/logger"
)

func main() {

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	log := logger.New()

	log.Info("Application started")

	log.Info("Configuration loaded")

	log.Info(
		"HTTP server", 
		"address", 
		cfg.HTTP.Address,
	)

	log.Info(
		"Database configuration",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
	)
}