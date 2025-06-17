package app

import (
	"context"

	grpcapp "github.com/zloyaloha/auth-service/internal/app/grpc"
	storage_app "github.com/zloyaloha/auth-service/internal/app/storage"
	"github.com/zloyaloha/auth-service/internal/config"
	"github.com/zloyaloha/auth-service/internal/services/auth"

	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
	StorageApp *storage_app.App
}

func New(
	logger *zap.Logger,
	config config.Config,
) *App {
	storageApp, err := storage_app.NewApp(context.TODO(), logger, config.PGConnectionConfig)

	if err != nil {
		logger.Error("Error while creating storage app", zap.Any("cfg", config.PGConnectionConfig))
		panic(err)
	}

	authService := auth.New(logger, storageApp.Storage(), storageApp.Storage(), config.TokenTTL)

	grpcApp := grpcapp.NewApp(logger, authService, config.GRPC.Port)

	return &App{
		GRPCServer: grpcApp,
		StorageApp: storageApp,
	}
}
