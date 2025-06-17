package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zloyaloha/auth-service/internal/domain/models"
)

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

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}