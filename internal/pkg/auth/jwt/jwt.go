package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maksemen2/medods-task/internal/pkg/auth"
	"time"
)

// jwtClaims имплементирует auth.Claims, payload jwt - Access токенов.
type jwtClaims struct {
	jwt.RegisteredClaims           // Встроенные зарегистрированные поля, будут использоваться exp, ID
	GUID                 uuid.UUID `json:"sub"`
	IP                   string    `json:"ip"`
}

// GetGUID - геттер для ID пользователя
func (c *jwtClaims) GetGUID() uuid.UUID {
	return c.GUID
}

// GetIP - геттер для айпи пользователя
func (c *jwtClaims) GetIP() string {
	return c.IP
}

// GetJTI - геттер для ID токена
func (c *jwtClaims) GetJTI() uuid.UUID {
	// Мы можем быть уверены, что ID спарсится, потому что всегда при создании токена мы кладем
	// в это поле валидный uuid. На клиенте это поле не может быть изменено,
	// а при утечке ключа JWT паника
	// при получении ID токена будет нашей меньшей проблемой.
	return uuid.MustParse(c.ID)
}

// JWTTokenManager имплементирует auth.AccessTokenManager.
type JWTTokenManager struct {
	SigningKey []byte        // Секретный ключ для подписи
	TokenTTL   time.Duration // Время истечения Access токена
}

// NewManager - конструктор JWTTokenManager.
// Принимает секретный ключ и время жизни токена.
func NewManager(signingKey []byte, tokenTTL time.Duration) auth.AccessTokenManager {
	return &JWTTokenManager{
		SigningKey: signingKey,
		TokenTTL:   tokenTTL,
	}
}

// Generate генерирует новый Access Token. Принимает GUID пользователя, ID токена и IP-адрес.
// Токен подписывается методом jwt.SigningMethodHS512 (SHA512) и секретным ключом.
// Возвращает строку с токеном.
func (m *JWTTokenManager) Generate(guid uuid.UUID, id uuid.UUID, ip string) (string, error) {
	currentTime := time.Now()

	claims := &jwtClaims{
		GUID: guid,
		IP:   ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id.String(),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(m.TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString(m.SigningKey)
}

// Parse парсит Access токен и возвращает его Claims.
// Возвращает ошибку, если токен невалиден или просрочен.
func (m *JWTTokenManager) Parse(raw string) (auth.Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, auth.ErrInvalidSignature
		}
		return m.SigningKey, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, auth.ErrTokenExpired
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return nil, auth.ErrInvalidSignature
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, auth.ErrTokenExpired
		}
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	return claims, nil
}
