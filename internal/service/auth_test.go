package service_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maksemen2/medods-task/internal/domain"
	"github.com/maksemen2/medods-task/internal/pkg/auth/crypto"
	mock_auth "github.com/maksemen2/medods-task/internal/pkg/auth/mocks"
	"github.com/maksemen2/medods-task/internal/pkg/auth/refresh"
	mock_repository "github.com/maksemen2/medods-task/internal/repository/mocks"
	"github.com/maksemen2/medods-task/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAuthService_AuthenticateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userRepo := mock_repository.NewMockIUserRepo(ctrl)
		tokenRepo := mock_repository.NewMockITokenRepo(ctrl)
		tokenManager := mock_auth.NewMockAccessTokenManager(ctrl)
		logger := zap.NewNop()
		svc := service.NewAuthServiceImpl(userRepo, tokenRepo, tokenManager, logger, time.Hour)

		guid := uuid.New()
		expectedAccessToken := "test_access"

		userRepo.EXPECT().Create(gomock.Any(), guid, gomock.Any()).Return(nil)
		tokenManager.EXPECT().Generate(guid, gomock.Any(), gomock.Any()).Return(expectedAccessToken, nil)
		tokenRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), guid, gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, tokenID, tokenJTI uuid.UUID, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) error {
				assert.NotEmpty(t, refreshTokenHash)
				return nil
			},
		)

		result, err := svc.AuthenticateUser(context.Background(), guid, "127.0.0.1")
		assert.NoError(t, err)
		assert.Equal(t, expectedAccessToken, result.AccessToken)

		decoded, err := base64.URLEncoding.DecodeString(result.RefreshToken)
		assert.NoError(t, err)
		assert.Len(t, decoded, refresh.TokenLength)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenRepo := mock_repository.NewMockITokenRepo(ctrl)
		tokenManager := mock_auth.NewMockAccessTokenManager(ctrl)
		claims := mock_auth.NewMockClaims(ctrl)
		logger := zap.NewNop()
		svc := service.NewAuthServiceImpl(nil, tokenRepo, tokenManager, logger, time.Hour)

		oldJTI := uuid.New()
		guid := uuid.New()
		oldRefresh := []byte("old_refresh")
		oldRefreshB64 := base64.URLEncoding.EncodeToString(oldRefresh)
		hashedOldRefresh, err := crypto.HashBytes(oldRefresh)
		assert.NoError(t, err)
		storedTokenID := uuid.New()

		tokenManager.EXPECT().Parse("valid_access").Return(claims, nil)
		claims.EXPECT().GetGUID().Return(guid)
		claims.EXPECT().GetJTI().Return(oldJTI)
		claims.EXPECT().GetIP().Return("old_ip")
		tokenRepo.EXPECT().GetToken(gomock.Any(), guid, oldJTI, gomock.Any()).Return(storedTokenID, hashedOldRefresh, nil)

		newAccessToken := "new_access"
		tokenManager.EXPECT().Generate(guid, gomock.Any(), "new_ip").Return(newAccessToken, nil)
		tokenRepo.EXPECT().RotateToken(
			gomock.Any(), storedTokenID, gomock.Any(), gomock.Any(), guid, gomock.Any(), gomock.Any(),
		).DoAndReturn(func(ctx context.Context, tokenID, newTokenID, newJTI uuid.UUID, userID uuid.UUID, hashedRefreshToken string, expiresAt time.Time) error {
			assert.NotEmpty(t, hashedRefreshToken)
			assert.GreaterOrEqual(t, len(hashedRefreshToken), 50)
			return nil
		})

		result, err := svc.RefreshToken(context.Background(), "valid_access", oldRefreshB64, "new_ip")
		assert.NoError(t, err)
		assert.Equal(t, newAccessToken, result.AccessToken)
		decoded, err := base64.URLEncoding.DecodeString(result.RefreshToken)
		assert.NoError(t, err)
		assert.Len(t, decoded, refresh.TokenLength)
	})

	t.Run("invalid access token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenManager := mock_auth.NewMockAccessTokenManager(ctrl)
		logger := zap.NewNop()
		svc := service.NewAuthServiceImpl(nil, nil, tokenManager, logger, time.Hour)

		tokenManager.EXPECT().Parse("invalid").Return(nil, errors.New("invalid"))

		_, err := svc.RefreshToken(context.Background(), "invalid", "refresh", "ip")
		assert.ErrorIs(t, err, domain.ErrInvalidAccessToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenManager := mock_auth.NewMockAccessTokenManager(ctrl)
		claims := mock_auth.NewMockClaims(ctrl)
		tokenRepo := mock_repository.NewMockITokenRepo(ctrl)
		logger := zap.NewNop()
		svc := service.NewAuthServiceImpl(nil, tokenRepo, tokenManager, logger, time.Hour)

		tokenManager.EXPECT().Parse("valid").Return(claims, nil)
		claims.EXPECT().GetGUID().Return(uuid.New())
		claims.EXPECT().GetJTI().Return(uuid.New())
		tokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(uuid.Nil, "wrong_hash", nil)

		_, err := svc.RefreshToken(
			context.Background(),
			"valid",
			base64.URLEncoding.EncodeToString([]byte("refresh")),
			"ip",
		)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("token not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tokenManager := mock_auth.NewMockAccessTokenManager(ctrl)
		claims := mock_auth.NewMockClaims(ctrl)
		tokenRepo := mock_repository.NewMockITokenRepo(ctrl)
		logger := zap.NewNop()
		svc := service.NewAuthServiceImpl(nil, tokenRepo, tokenManager, logger, time.Hour)

		tokenManager.EXPECT().Parse("valid").Return(claims, nil)
		claims.EXPECT().GetGUID().Return(uuid.New())
		claims.EXPECT().GetJTI().Return(uuid.New())
		tokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(uuid.Nil, "", domain.ErrTokenNotFound)

		_, err := svc.RefreshToken(
			context.Background(),
			"valid",
			base64.URLEncoding.EncodeToString([]byte("refresh")),
			"ip",
		)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}
