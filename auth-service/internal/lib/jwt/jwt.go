package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zloyaloha/auth-service/internal/domain/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// Claims структура для JWT claims
type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	AppID  int    `json:"app_id"`
	Role   string `json:"urole"`
	jwt.RegisteredClaims
}

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	if app.Secret == "" {
		return "", fmt.Errorf("empty app secret")
	}

	token := jwt.New(jwt.SigningMethodHS256) // TODO: replace to rsa

	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID
	claims["urole"] = user.Role

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string, appSecret string) (*Claims, error) {
	if appSecret == "" {
		return nil, fmt.Errorf("empty app secret")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(appSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("%w", ErrInvalidToken)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("%w", ErrTokenExpired)
	}

	return claims, nil
}

func ValidateTokenWithAppID(tokenString string, appSecret string, expectedAppID int) (*Claims, error) {
	claims, err := ValidateToken(tokenString, appSecret)
	if err != nil {
		return nil, err
	}

	if claims.AppID != expectedAppID {
		return nil, fmt.Errorf("invalid app_id: expected %d, got %d", expectedAppID, claims.AppID)
	}

	return claims, nil
}
