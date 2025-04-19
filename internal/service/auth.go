package service

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maksemen2/medods-task/internal/domain"
	"github.com/maksemen2/medods-task/internal/pkg/auth"
	"github.com/maksemen2/medods-task/internal/pkg/auth/crypto"
	"github.com/maksemen2/medods-task/internal/pkg/auth/refresh"
	"github.com/maksemen2/medods-task/internal/repository"
	"go.uber.org/zap"
	"time"
)

// IAuthService - интерфейс для работы с аутентификацией пользователей.
type IAuthService interface {
	AuthenticateUser(ctx context.Context, guid uuid.UUID, ip string) (*domain.UserAuth, error)
	RefreshToken(ctx context.Context, accessToken, refreshToken, ip string) (*domain.UserAuth, error)
}

type AuthServiceImpl struct {
	userRepo     repository.IUserRepo
	tokenRepo    repository.ITokenRepo
	tokenManager auth.AccessTokenManager
	logger       *zap.Logger
	refreshTTL   time.Duration
}

func NewAuthServiceImpl(userRepo repository.IUserRepo, tokenRepo repository.ITokenRepo, tokenManager auth.AccessTokenManager, logger *zap.Logger, refreshTTL time.Duration) IAuthService {
	return &AuthServiceImpl{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenManager: tokenManager,
		logger:       logger,
		refreshTTL:   refreshTTL,
	}
}

// Функция - заглушка, которая отправляет уведомление пользователю об обновлении токена с другого айпи - адреса.
// Вероятно, для этого лучше использовать отдельный сервис, но, так как это просто заглушка,
// я сделал это в виде метода сервиса аутентификации.
func (s *AuthServiceImpl) notifyUser(guid uuid.UUID, oldIP, newIP string) error {
	// Логика получения почты пользователя и отправки уведомления
	return nil
}

// AuthenticateUser - аутентификация пользователя по guid.
// Возвращает доменную модель domain.UserAuth.
func (s *AuthServiceImpl) AuthenticateUser(ctx context.Context, guid uuid.UUID, ip string) (*domain.UserAuth, error) {
	err := s.userRepo.Create(ctx, guid, gofakeit.Email()) // Используется моковая почта
	// Если пользователь уже существует - для нас это не проблема, просто идем дальше
	if err != nil && !errors.Is(err, domain.ErrUserExists) {
		return nil, domain.ErrUnexpected
	}

	jti := uuid.New()

	accessToken, err := s.tokenManager.Generate(guid, jti, ip)

	if err != nil {
		s.logger.Error("Error generating token", zap.Error(err))
		return nil, domain.ErrUnexpected
	}

	refreshToken, err := refresh.GenerateToken()
	if err != nil {
		s.logger.Error("Error generating refresh token", zap.Error(err))
		return nil, domain.ErrUnexpected
	}

	refreshTokenHash, err := crypto.HashBytes(refreshToken)
	if err != nil {
		s.logger.Error("Error hashing refresh token", zap.Error(err))
		return nil, domain.ErrUnexpected
	}

	err = s.tokenRepo.Create(ctx, uuid.New(), jti, guid, refreshTokenHash, time.Now().Add(s.refreshTTL)) // Записываем обязательно хешированный токен

	if err != nil {
		if errors.Is(err, domain.ErrTokenExists) {
			// Мы не можем допускать ситуации, когда уже существует Refresh токен для пары пользователя и Access - токена
			// Поэтому логируем это как ошибку
			s.logger.Error("Token already exists", zap.String("guid", guid.String()), zap.String("jti", jti.String()))
		}
		return nil, domain.ErrUnexpected
	}

	return &domain.UserAuth{
		AccessToken:  accessToken,
		RefreshToken: base64.URLEncoding.EncodeToString(refreshToken),
	}, nil
}

// RefreshToken обновляет токены пользователя.
// Проверяет валидность Access токена и Refresh токена.
// Если токены валидны, генерирует новые токены и обновляет Refresh - токен в базе данных.
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, accessToken, refreshToken, ip string) (*domain.UserAuth, error) {
	claims, err := s.tokenManager.Parse(accessToken)
	if err != nil {
		s.logger.Debug("Bad token provided", zap.Error(err))
		return nil, domain.ErrInvalidAccessToken
	}

	refreshTokenBytes, err := base64.URLEncoding.DecodeString(refreshToken)
	if err != nil {
		s.logger.Debug("Bad refresh token provided", zap.Error(err))
		return nil, domain.ErrInvalidRefreshToken
	}

	guid := claims.GetGUID()
	jti := claims.GetJTI()
	currentTime := time.Now()
	storedTokenID, storedTokenHash, err := s.tokenRepo.GetToken(ctx, guid, jti, currentTime)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			s.logger.Debug("Token not found", zap.String("guid", guid.String()), zap.String("jti", jti.String()))
			return nil, err
		}
		return nil, domain.ErrUnexpected
	}

	if !crypto.CompareHashAndBytes(refreshTokenBytes, storedTokenHash) {
		s.logger.Debug("Invalid refresh token provided", zap.String("guid", guid.String()), zap.String("jti", jti.String()))
		return nil, domain.ErrInvalidRefreshToken
	}

	newJTI := uuid.New()
	newAccessToken, err := s.tokenManager.Generate(guid, newJTI, ip)
	if err != nil {
		s.logger.Error("Error generating new token", zap.Error(err))
		return nil, domain.ErrUnexpected
	}

	newRefreshToken, err := refresh.GenerateToken()
	if err != nil {
		s.logger.Error("Error generating new refresh token", zap.Error(err))
		return nil, domain.ErrUnexpected
	}

	hashedRefreshToken, err := crypto.HashBytes(newRefreshToken)
	if err != nil {
		s.logger.Error("Error hashing new refresh token", zap.Error(err))
		return nil, domain.ErrUnexpected
	}

	expirationTime := currentTime.Add(s.refreshTTL)
	err = s.tokenRepo.RotateToken(ctx, storedTokenID, uuid.New(), newJTI, guid, hashedRefreshToken, expirationTime)
	if err != nil {
		return nil, domain.ErrUnexpected
	}

	if oldIP := claims.GetIP(); oldIP != ip {
		go s.notifyUser(guid, oldIP, ip)
	}

	return &domain.UserAuth{
		AccessToken:  newAccessToken,
		RefreshToken: base64.URLEncoding.EncodeToString(newRefreshToken),
	}, nil
}
