package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/zloyaloha/auth-service/internal/domain/models"
	"github.com/zloyaloha/auth-service/internal/lib/jwt"
	storage "github.com/zloyaloha/auth-service/internal/storage/users-storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserStorage interface {
	SaveUser(ctx context.Context, email, first_name, last_name string, hash []byte) (int64, error)
	GetUser(ctx context.Context, email string) (*models.User, error)
	GetRole(ctx context.Context, user_id int64) (string, error)
}

type AppProvider interface {
	SaveApp(ctx context.Context, name, secret string) error
	GetApp(ctx context.Context, appID int) (*models.App, error)
}

type Auth struct {
	logger      *zap.Logger
	usrStorage  UserStorage
	appProvider AppProvider
	tokenTTL    time.Duration
}

func New(
	logger *zap.Logger,
	userStorage UserStorage,
	appProvider AppProvider,
	ttl time.Duration,
) *Auth {
	return &Auth{
		logger:      logger,
		usrStorage:  userStorage,
		appProvider: appProvider,
		tokenTTL:    ttl,
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
			return "", fmt.Errorf("error: %w", ErrInvalidCredentials)
		}

		a.logger.Error("failed to get user", zap.Error(err), zap.String("email", email))
		return "", fmt.Errorf("error: %w", err)
	}

	role, err := a.usrStorage.GetRole(ctx, user.ID)
	if err != nil {
		a.logger.Error("failed to get user role", zap.Error(err), zap.Int64("email", user.ID))
		return "", fmt.Errorf("error: %w", err)
	}
	user.Role = role

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid password", zap.String("email", email))
		return "", fmt.Errorf("error: %w", ErrInvalidCredentials)
	}

	app, err := a.appProvider.GetApp(ctx, appID)
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

func (a *Auth) ValidateToken(ctx context.Context, tokenString string) (*jwt.Claims, error) {
	a.logger.Info("validating token")

	app, err := a.appProvider.GetApp(ctx, 1) // TODO
	if err != nil {
		a.logger.Error("failed to get app", zap.Error(err))
		return nil, err
	}

	claims, err := jwt.ValidateToken(tokenString, app.Secret)
	if err != nil {
		a.logger.Error("invalid token", zap.Error(err))
		return nil, err
	}

	return claims, nil
}

func (a *Auth) GetRole(ctx context.Context, userID string) (string, error) {
	a.logger.Info("checking user role", zap.String("userID", userID))

	requestedUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		a.logger.Warn("invalid user_id format", zap.String("userID", userID), zap.Error(err))
		return "", fmt.Errorf("invalid user_id format: %w", err)
	}

	role, err := a.usrStorage.GetRole(ctx, requestedUserID)
	if err != nil {
		a.logger.Warn("failed to get user role", zap.Error(err), zap.Int64("userID", requestedUserID))
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	a.logger.Info("role check completed", zap.Int64("userID", requestedUserID), zap.String("role", role))

	return role, nil
}