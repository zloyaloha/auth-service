package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/zloyaloha/auth-service/internal/domain/models"
	"github.com/zloyaloha/auth-service/internal/lib/jwt"
	"github.com/zloyaloha/auth-service/internal/storage/users-storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserStorage interface {
	SaveUser(ctx context.Context, email, first_name, last_name string, hash []byte) (int64, error)
	GetUser(ctx context.Context, email string) (*models.User, error)
}

type AppProvider interface {
	SaveApp(ctx context.Context, name, secret string) error
	GetApp(ctx context.Context, appID int) (*models.App, error)
}

type Auth struct {
	logger *zap.Logger
	usrStorage UserStorage
	appProvider AppProvider
	tokenTTL time.Duration
}

func New(
	logger *zap.Logger,
	userStorage UserStorage,
	appProvider AppProvider,
	ttl time.Duration,
) *Auth {
	return &Auth{
		logger: logger,
		usrStorage: userStorage,
		appProvider: appProvider,
		tokenTTL: ttl,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, first_name, last_name string, password string) (int64, error) {
	a.logger.Info("register new user", zap.String("email", email), zap.String("first_name", first_name), zap.String("last_name", last_name))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // тут сразу генерируется и соль и хэш
	if err != nil {
		a.logger.Error("failed to generate password hash", zap.Error(err))
		return 0, fmt.Errorf("error: %w", err)
	}

	user_id, err := a.usrStorage.SaveUser(ctx, email, first_name, last_name, hash)

	if err != nil {
		a.logger.Error("Failed to save user to DB", zap.Error(err))
		return 0, fmt.Errorf("error: %w", err)
	}

	return user_id, nil
}

func (a *Auth) Login(
	ctx context.Context,
	email, password string,
	appID int,
) (string, error) {
	a.logger.Info("attempting to log in user", zap.String("email", email))

	user, err := a.usrStorage.GetUser(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.logger.Warn("user not found", zap.String("email", email))
			return "", fmt.Errorf("error: %w", err)
		}

		a.logger.Error("failed to get user", zap.Error(err), zap.String("email", email))
		return "", fmt.Errorf("error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid credentials", zap.String("email", email))
		return "", fmt.Errorf("error: %w", err)
	}

	app, err := a.appProvider.GetApp(ctx, appID);
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.logger.Warn("app not found", zap.Int("appID", appID))
			return "", fmt.Errorf("error: %w", err)
		}
		a.logger.Error("failed to get app", zap.Error(err), zap.Int("appID", appID))
		return "", fmt.Errorf("error: %w", err)
	}

	a.logger.Info("user succesfully logged in")

	token, err := jwt.NewToken(*user, *app, a.tokenTTL)

	if err != nil {
		a.logger.Error("failed to generate token", zap.Error(err))
		return "", fmt.Errorf("error: %w", err)
	}

	return token, nil
}