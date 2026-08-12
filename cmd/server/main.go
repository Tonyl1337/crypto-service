package main

import (
	"context"
	"log"

	"github.com/Tonyl1337/crypto-service/internal/app"
)

func main() {
	ctx := context.Background()

	app, err := app.New("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
