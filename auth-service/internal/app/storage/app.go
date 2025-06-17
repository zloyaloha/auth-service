package storage_app

import (
	"context"

	"github.com/zloyaloha/auth-service/internal/config"
	storage "github.com/zloyaloha/auth-service/internal/storage/users-storage"
	"go.uber.org/zap"
)

type App struct {
	storage storage.Storage
	logger  *zap.Logger
}

func NewApp(
	ctx context.Context,
	logger *zap.Logger,
	config config.PGConnectionConfig,
) (*App, error) {
	strg, err := storage.NewPGHandler(ctx, config)

	if err != nil {
		return nil, err
	}

	return &App{
		storage: strg,
		logger:  logger,
	}, nil
}

func (a *App) Storage() storage.Storage {
	return a.storage
}

func (a *App) Stop() {
	a.logger.Info("Storage app stopping")

	a.storage.Stop()
}
