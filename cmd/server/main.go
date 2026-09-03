package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tonyl1337/crypto-service/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(ctx); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	log.Println("application stopped")
}
