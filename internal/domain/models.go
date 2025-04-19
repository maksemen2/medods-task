package domain

// UserAuth - доменная модель для хранения и передачи данных аутентификации пользователя.
type UserAuth struct {
	AccessToken  string // Токен доступа
	RefreshToken string // Refresh - токен, закодированный в base64
}
