package postgresqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/maksemen2/medods-task/internal/domain"
	"github.com/maksemen2/medods-task/internal/pkg/database"
	"github.com/maksemen2/medods-task/internal/repository"
	"go.uber.org/zap"
)

// PostgresqlUserRepo - имплементация интерфейса repository.IUserRepo.
// Позволяет взаимодействовать с сущностями пользователя в Postgresql
type PostgresqlUserRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// Create создает нового пользователя с указанными guid и email.
// Если пользователь с таким guid уже существует, возвращает ошибку domain.ErrUserExists.
func (r *PostgresqlUserRepo) Create(ctx context.Context, guid uuid.UUID, email string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (guid, email) VALUES ($1, $2)", guid, email)
	if err != nil {
		if database.IsPGError(err, database.PGUniqueViolationCode) {
			return domain.ErrUserExists
		}
		r.logger.Error("Error inserting user", zap.Error(err))
		return err
	}
	return nil
}

// GetEmail возвращает email пользователя по его guid.
// Если пользователь не найден, возвращает ошибку domain.ErrUserNotFound.
func (r *PostgresqlUserRepo) GetEmail(ctx context.Context, guid uuid.UUID) (string, error) {
	var email string

	err := r.db.GetContext(ctx, &email, "SELECT email FROM users WHERE guid = $1", guid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		r.logger.Error("Error querying user", zap.Error(err))
		return "", err
	}

	return email, nil
}

// NewPostgresqlUserRepo - конструктор для создания нового экземпляра PostgresqlUserRepo.
func NewPostgresqlUserRepo(db *sqlx.DB, logger *zap.Logger) repository.IUserRepo {
	return &PostgresqlUserRepo{
		db:     db,
		logger: logger,
	}
}
