package grpcapp

import (
	"fmt"
	"net"
	"time"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"github.com/zloyaloha/auth-service/internal/grpc/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type App struct {
	logger *zap.Logger
	gRPCServer *grpc.Server
	port int
}

func NewApp(logger *zap.Logger, authService auth.Auth, port int) (*App) {
	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}

	grpc_zap.ReplaceGrpcLoggerV2(logger)

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(logger, opts...),
		),
	)

	auth.Register(gRPCServer, authService)

	return &App{
		logger: logger,
		gRPCServer: gRPCServer,
		port: port,
	}
}

func (a* App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a* App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("error while staring tcp server: %w", err)
	}

	defer l.Close()

	a.logger.Info("grpc server started", zap.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("error while serving grpc server: %w", err)
	}

	return nil;
}