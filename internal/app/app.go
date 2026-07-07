package app

import (
	"log/slog"

	"github.com/Tonyl1337/crypto-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *pgxpool.Pool
}

func New(cfg *config.Config, logger *slog.Logger, db *pgxpool.Pool,) *App {

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}
}