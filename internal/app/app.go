package app

import (
	"context"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/config"
	"github.com/Tonyl1337/crypto-service/internal/database"
	"github.com/Tonyl1337/crypto-service/internal/repository/postgres"
	"github.com/Tonyl1337/crypto-service/internal/service"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest/handler"
)

type App struct {
	server *rest.Server
}

func New(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	rateRepo := postgres.NewRateRepository(db)

	coinClient := coingecko.NewClient()

	rateService := service.NewRateService(
		rateRepo,
		coinClient,
	)

	rateHandler := handler.NewRateHandler(rateService)

	server := rest.NewServer(
		cfg.HTTP.Address,
		rateHandler,
	)

	return &App{
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.server.Run(ctx)
}
