package repository

import (
	"context"
	"github.com/google/uuid"
)

// IUserRepo - интерфейс для работы с сущностями пользователей в базе данных
type IUserRepo interface {
	Create(ctx context.Context, guid uuid.UUID, email string) error // Create создает нового пользователя с указанными guid и email
	GetEmail(ctx context.Context, guid uuid.UUID) (string, error)   // GetEmail возвращает email пользователя по его guid
}
