package crypto_test

import (
	"github.com/maksemen2/medods-task/internal/pkg/auth/crypto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestHashBytes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		input := []byte("best_refresh_token")
		hash, err := crypto.HashBytes(input)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("Input Too Long", func(t *testing.T) {
		input := []byte("a" + string(make([]byte, 72)))
		hash, err := crypto.HashBytes(input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
		assert.Empty(t, hash)
	})
}

func TestCompareHashAndBytes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		input := []byte("best_refresh_token")
		hash, err := crypto.HashBytes(input)
		assert.NoError(t, err)

		result := crypto.CompareHashAndBytes(input, hash)
		assert.True(t, result)
	})

	t.Run("Fail", func(t *testing.T) {
		input := []byte("best_refresh_token")
		hash := "bad_hash"

		result := crypto.CompareHashAndBytes(input, hash)
		assert.False(t, result)
	})
}
