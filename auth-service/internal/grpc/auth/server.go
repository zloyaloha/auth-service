package auth

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/zloyaloha/auth-service/internal/lib/jwt"
	"github.com/zloyaloha/auth-service/internal/services/auth"
	storage "github.com/zloyaloha/auth-service/internal/storage/users-storage"
	ssov1 "github.com/zloyaloha/protos/gen/go/sso"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		first_name string,
		last_name string,
		password string,
	) (int64, error)

	ValidateToken(
		ctx context.Context,
		tokenString string,
	) (*jwt.Claims, error)

	GetRole(
		ctx context.Context,
		user_id string,
	) (string, error)
}

func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userID, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetFirstName(), in.GetLastName(), in.GetPassword())

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to registrate")
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

// extractTokenFromContext извлекает токен из заголовка Authorization
func extractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata found")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.New("no authorization header found")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

func (s *serverAPI) GetRole(
	ctx context.Context,
	in *ssov1.GetRoleRequest,
) (*ssov1.GetRoleResponse, error) {
	if in.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	token, err := extractTokenFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization header")
	}

	claims, err := s.auth.ValidateToken(ctx, token)
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, status.Error(codes.Unauthenticated, "token expired")
		}
		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	if claims.Role != "admin" {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	role, err := s.auth.GetRole(ctx, in.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get role")
	}

	return &ssov1.GetRoleResponse{Role: role}, nil
}
