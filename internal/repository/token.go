package repository

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// ITokenRepo - интерфейс для работы с сущностями Refresh токенов в базе данных
type ITokenRepo interface {
	Create(ctx context.Context, id, jti, userID uuid.UUID, token string, expiresAt time.Time) error             // Create создает новый Refresh - токен
	GetToken(ctx context.Context, userID, jti uuid.UUID, notAfter time.Time) (uuid.UUID, string, error)         // GetToken получает Refresh - токен из базы данных. Возвращает айди токена и токен.
	RotateToken(ctx context.Context, oldID, id, jti, userID uuid.UUID, token string, expiresAt time.Time) error // RotateToken производит ротацию токена, т.е. удаление старого и создание нового.
}
