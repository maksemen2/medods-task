package domain

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrTokenNotFound       = errors.New("token not found")
	ErrTokenExists         = errors.New("token already exists")
	ErrUnexpected          = errors.New("unexpected error")
	ErrInvalidRefreshToken = errors.New("invalid refresh token provided")
	ErrInvalidAccessToken  = errors.New("invalid access token provided")
)
