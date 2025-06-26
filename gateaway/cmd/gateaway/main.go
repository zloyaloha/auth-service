package main

import (
	"context"

	"github.com/zloyaloha/gateaway/internal/config"
	"github.com/zloyaloha/gateaway/internal/http-server/app"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
    config := config.MustLoad()
    logger := buildLogger(config.Env)

    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    app := app.New(ctx, logger, config)
    app.HTTPServer.MustRun()
}

func buildLogger(env string) *zap.Logger {
	switch (env) {
	case envLocal:
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err := config.Build()
		if err != nil {
			panic("can't build logger")
		}
		return logger
	case envProd:
		config := zap.NewProductionConfig()
		logger, err := config.Build()
		if err != nil {
			panic("can't build logger")
		}
		return logger
	default:
		panic("unknown env")
	}
}