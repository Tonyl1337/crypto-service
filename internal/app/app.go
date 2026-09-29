package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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
	db                 *pgxpool.Pool
}

func New() (*App, error) {

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	rateRepo := postgres.NewRateRepository(db)

	subscriptionRepo := postgres.NewSubscriptionRepository(db)

	subscriptionService := service.NewSubscriptionsService(
		subscriptionRepo,
	)

	coinClient := coingecko.NewClient()

	rateService := service.NewRateService(
		rateRepo,
		coinClient,
	)

	telegramBot, err := telegram.NewBot(cfg.Telegram.Token)
	if err != nil {
		return nil, err
	}

	telegramHandler := telegram.NewHandler(
		telegramBot,
		rateService,
		subscriptionService,
		coinClient,
	)

	subscriptionSender := scheduler.NewSubscriptionSender(
		subscriptionService,
		rateService,
		telegramBot,
		10*time.Second,
	)

	rateHandler := handler.NewRateHandler(
		rateService,
		coinClient,
	)

	httpAddress := fmt.Sprintf(
		"%s:%s",
		cfg.HTTP.Host,
		cfg.HTTP.Port,
	)

	server := rest.NewServer(
		httpAddress,
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
		db:                 db,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.db.Close()

	a.updater.Start(ctx)

	a.subscriptionSender.Start(ctx)

	a.bot.Start(
		ctx,
		a.telegramHandler.Handle,
	)

	return a.server.Run(ctx)
}
