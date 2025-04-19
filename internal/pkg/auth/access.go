package auth

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidSignature = errors.New("invalid signature")
)

// Claims описывает payload Access токенов.
type Claims interface {
	GetGUID() uuid.UUID // GetGUID возвращает ID пользователя из Claims токена.
	GetIP() string      // GetIP возвращает IP-адрес из Claims токена
	GetJTI() uuid.UUID  // GetJTI возвращает ID токена из Claims токена
}

// AccessTokenManager описывает интерфейс менеджера Access токенов.
type AccessTokenManager interface {
	Generate(guid uuid.UUID, id uuid.UUID, ip string) (string, error) // Generate генерирует новый AccessToken для пользователя с добавлением его ip-адреса и айди токена.
	Parse(raw string) (Claims, error)                                 // Parse парсит AccessToken и возвращает его Claims
}
