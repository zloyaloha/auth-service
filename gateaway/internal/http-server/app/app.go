package app

import (
	"context"

	"github.com/zloyaloha/gateaway/internal/config"
	httpapp "github.com/zloyaloha/gateaway/internal/http-server/app/http"
	"go.uber.org/zap"
)

type App struct {
	HTTPServer *httpapp.App
}

func New(
	ctx context.Context,
	logger *zap.Logger,
	config config.Config,
) *App {
	app := httpapp.New(ctx, logger, config.HTTP.Address, config.HTTP.AuthServiceAddress)

	return &App{
		HTTPServer: app,
	}
}