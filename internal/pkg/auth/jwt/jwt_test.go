package jwt_test

import (
	"github.com/google/uuid"
	"github.com/maksemen2/medods-task/internal/pkg/auth"
	"github.com/maksemen2/medods-task/internal/pkg/auth/jwt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJWTTokenManager_GenerateAndParse(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		manager := jwt.NewManager([]byte("very_secret_key"), 10*time.Minute)

		guid := uuid.New()
		ip := "127.0.0.1"

		token, err := manager.Generate(guid, uuid.New(), ip)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.Parse(token)

		assert.NoError(t, err)
		assert.NotEmpty(t, claims)

		assert.Equal(t, claims.GetGUID(), guid)
		assert.Equal(t, claims.GetIP(), ip)
	})

	t.Run("Token Expired", func(t *testing.T) {
		manager := jwt.NewManager([]byte("very_secret_key"), 1*time.Millisecond)

		token, err := manager.Generate(uuid.New(), uuid.New(), "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		time.Sleep(2 * time.Millisecond)

		claims, err := manager.Parse(token)
		assert.Error(t, err)
		assert.ErrorIs(t, err, auth.ErrTokenExpired)
		assert.Empty(t, claims)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		manager := jwt.NewManager([]byte("very_secret_key"), 10*time.Minute)

		token, err := manager.Generate(uuid.New(), uuid.New(), "127.0.0.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		token = token[:len(token)/2] + "A" + token[len(token)/2+1:]

		claims, err := manager.Parse(token)
		assert.Error(t, err)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Empty(t, claims)
	})
}
