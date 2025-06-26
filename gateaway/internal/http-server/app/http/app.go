package httpapp

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	mvcors "github.com/zloyaloha/gateaway/internal/http-server/middleware/cors"
	mvlogger "github.com/zloyaloha/gateaway/internal/http-server/middleware/logger"
	ssov1 "github.com/zloyaloha/protos/gen/go/sso"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

type App struct {
	logger *zap.Logger
	address string
	handler http.Handler
}

func New(ctx context.Context, logger *zap.Logger, address, authServiceAddress string) *App {
	mux := runtime.NewServeMux()

    cc, err := grpc.NewClient(authServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

    if err != nil {
		logger.Error("failed to dial grpc server", zap.Error(err))
    }

    err = ssov1.RegisterAuthHandlerClient(ctx, mux, ssov1.NewAuthClient(cc))
    if err != nil {
		logger.Error("failed to register grpc-gateway client", zap.Error(err))
    }

    handler := mvcors.New(mux)
    handler = mvlogger.New(handler, logger)

	return &App{
		logger: logger,
		address: address,
		handler: handler,
	}
}

func (a *App) MustRun() {
	a.logger.Info("server started on", zap.String("addr", a.address))
    if err := http.ListenAndServe(a.address, a.handler); err != nil {
		a.logger.Error("failed to serve", zap.Error(err))
    }
}