package app

import (
	"context"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/config"
	"github.com/Tonyl1337/crypto-service/internal/database"
	"github.com/Tonyl1337/crypto-service/internal/repository/postgres"
	"github.com/Tonyl1337/crypto-service/internal/scheduler"
	"github.com/Tonyl1337/crypto-service/internal/service"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest/handler"
	"github.com/Tonyl1337/crypto-service/internal/transport/telegram"
)

type App struct {
	server             *rest.Server
	updater            *scheduler.Updater
	subscriptionSender *scheduler.SubscriptionSender
	bot                *telegram.Bot
	telegramHandler    *telegram.Handler
}

func New() (*App, error) {

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return nil, err
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	rateRepo := postgres.NewRateRepository(db)

	subscriptionRepo := postgres.NewSubscriptionRepository(db)

	coinClient := coingecko.NewClient()

	rateService := service.NewRateService(
		rateRepo,
		coinClient,
	)

	subscriptionService := service.NewSubscriptionsService(
		subscriptionRepo,
	)

	telegramBot, err := telegram.NewBot(cfg.Telegram.Token)
	if err != nil {
		return nil, err
	}

	telegramHandler := telegram.NewHandler(
		telegramBot,
		rateService,
		subscriptionService,
	)

	subscriptionSender := scheduler.NewSubscriptionSender(
		subscriptionService,
		rateService,
		telegramBot,
		10*time.Second,
	)

	rateHandler := handler.NewRateHandler(rateService)

	server := rest.NewServer(
		cfg.HTTP.Address,
		rateHandler,
	)

	updater := scheduler.NewUpdater(
		rateService,
		cfg.Scheduler.Interval,
	)

	return &App{
		server:             server,
		updater:            updater,
		subscriptionSender: subscriptionSender,
		bot:                telegramBot,
		telegramHandler:    telegramHandler,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.updater.Start(ctx)

	a.subscriptionSender.Start(ctx)

	a.bot.Start(
		ctx,
		a.telegramHandler.Handle,
	)

	return a.server.Run(ctx)
}
