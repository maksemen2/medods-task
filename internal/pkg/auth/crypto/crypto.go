package crypto

import "golang.org/x/crypto/bcrypt"

// HashBytes хеширует массив байт с помощью bcrypt и возвращает хеш в виде строки.
// Использует стандартную стоимость хэширования bcrypt (bcrypt.DefaultCost (10)).
// Если произошла ошибка, возвращает ее.
func HashBytes(plain []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(plain, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareHash сравнивает байты обычной строки с хенированной строки.
func CompareHashAndBytes(plain []byte, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), plain) == nil
}
