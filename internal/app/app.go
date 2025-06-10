package app

import (
	"context"
	"time"

	grpcapp "github.com/zloyaloha/auth-service/internal/app/grpc"
	"github.com/zloyaloha/auth-service/internal/services/auth"
	"github.com/zloyaloha/auth-service/internal/storage/users-storage"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	logger *zap.Logger,
	grpcPort int,
	tokenTTL time.Duration,
) *App {
	storage, err := storage.NewPGHandler(context.TODO())

	if err != nil {
		panic(err)
	}

	authService := auth.New(logger, storage, storage, tokenTTL)

	grpcApp := grpcapp.NewApp(logger, authService, grpcPort)

	return &App {
		GRPCServer: grpcApp,
	}
}