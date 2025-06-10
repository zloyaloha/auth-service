package main

import (
	"github.com/zloyaloha/auth-service/internal/app"
	"github.com/zloyaloha/auth-service/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	logger := buildLogger(cfg.Env)

	app := app.New(logger, cfg.GRPC.Port, cfg.TokenTTL)

	app.GRPCServer.MustRun()
}

func buildLogger(env string) *zap.Logger {
	var logger *zap.Logger

	switch env {
	case envLocal:
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = config.Build()
	case envDev:
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		logger, _ = config.Build()
	case envProd:
		logger, _ = zap.NewProduction()
	}

	return logger
}