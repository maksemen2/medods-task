package refresh

import (
	"crypto/rand"
)

const TokenLength = 64

// GenerateToken создает новый Refresh токен путем генерации случайной строки длиной 64 символа.
// Возвращается токен в виде байтового массива для
// избежания дублирования операций преобразования в и из строки
func GenerateToken() ([]byte, error) {
	raw := make([]byte, TokenLength)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}

	return raw, nil
}
