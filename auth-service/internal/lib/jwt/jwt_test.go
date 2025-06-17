package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/zloyaloha/auth-service/internal/domain/models"
)

var (
	testUser = models.User{
		ID: 123,
		Email: "test@example.com",
	}

	testApp = models.App{
		ID: 1,
		Secret: "test_secret_key_123",
	}

	testDuration = time.Hour

	testAppEmptySecret = models.App{
		ID: 2,
		Secret: "",
	}
)

func TestNewToken(t *testing.T) {
	t.Run("TestNewTokenCreating", func(t *testing.T) {
		token, err := NewToken(testUser, testApp, testDuration)

		assert.Nil(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, len(token), 100)
	})

	t.Run("TestNewToken_Claims", func(t *testing.T) {
		tokenString, err := NewToken(testUser, testApp, testDuration)

		assert.NoError(t, err)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(testApp.Secret), nil
		})

		assert.NoError(t, err)
		claims := token.Claims.(jwt.MapClaims)

		assert.Equal(t, float64(testUser.ID), claims["uid"])
		assert.Equal(t, testUser.Email, claims["email"])
		assert.Equal(t, float64(testApp.ID), claims["app_id"])

		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		assert.True(t, exp.After(time.Now()))
		assert.True(t, exp.Before(time.Now().Add(testDuration + time.Minute)))
	})

	t.Run("TestNewToken_EmptySecret", func(t *testing.T) {
		_, err := NewToken(testUser, testAppEmptySecret, testDuration)

		assert.Error(t, err)
	})
}