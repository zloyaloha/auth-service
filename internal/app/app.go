package app

import (
	"context"

	grpcapp "github.com/zloyaloha/auth-service/internal/app/grpc"
	"github.com/zloyaloha/auth-service/internal/config"
	"github.com/zloyaloha/auth-service/internal/services/auth"
	"github.com/zloyaloha/auth-service/internal/storage/users-storage"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	logger *zap.Logger,
	config config.Config,
) *App {
	storage, err := storage.NewPGHandler(context.TODO(), config.PGConnectionConfig)

	if err != nil {
		panic(err)
	}

	authService := auth.New(logger, storage, storage, config.TokenTTL)

	grpcApp := grpcapp.NewApp(logger, authService, config.GRPC.Port)

	return &App {
		GRPCServer: grpcApp,
	}
}