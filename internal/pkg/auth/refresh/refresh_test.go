package refresh_test

import (
	"github.com/maksemen2/medods-task/internal/pkg/auth/refresh"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		token, err := refresh.GenerateToken()

		assert.NoError(t, err)
		assert.Len(t, token, refresh.TokenLength)
	})

}
