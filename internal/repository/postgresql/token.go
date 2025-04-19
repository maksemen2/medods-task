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
	"time"
)

// PostgresqlTokenRepo - имплементация интерфейса repository.ITokenRepo.
// Позволяет взаимодействовать с сущностями Refresh - токена в Postgresql
type PostgresqlTokenRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// Create создает новую запись о Refresh - токене с указанными id, jti, userID и token.
func (r *PostgresqlTokenRepo) Create(ctx context.Context, id, jti, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO tokens (id, jti, user_id, token, expires_at) VALUES ($1, $2, $3, $4, $5)", id, jti, userID, token, expiresAt)
	if err != nil {
		if database.IsPGError(err, database.PGUniqueViolationCode) {
			return domain.ErrTokenExists
		}
		r.logger.Error("error creating token", zap.Error(err))
		return err
	}
	return nil
}

// GetToken получает Refresh - токен из базы данных по userID и jti.
// Так же принимает notAfter - время, до которого токен должен быть действителен.
// Если токен не найден или просрочен, возвращает ошибку domain.ErrTokenNotFound.
func (r *PostgresqlTokenRepo) GetToken(ctx context.Context, userID, jti uuid.UUID, notAfter time.Time) (uuid.UUID, string, error) {
	var tokenID uuid.UUID
	var token string

	err := r.db.QueryRowxContext(ctx, "SELECT id, token FROM tokens WHERE user_id = $1 AND jti = $2 AND expires_at >= $3", userID, jti, notAfter).Scan(&tokenID, &token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, "", domain.ErrTokenNotFound
		}
		return uuid.Nil, "", err
	}

	return tokenID, token, nil
}

// RotateToken - обновляет токен в базе данных.
// Удаляет старый токен по oldID и создает новый с указанными параметрами.
// Если новый токен уже существует, возвращает ошибку domain.ErrTokenExists.
func (r *PostgresqlTokenRepo) RotateToken(ctx context.Context, oldID, id, jti, userID uuid.UUID, token string, expiresAt time.Time) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("error starting transaction", zap.Error(err))
		return err
	}
	defer database.TxRollback(tx, r.logger)

	_, err = tx.ExecContext(ctx, "DELETE FROM tokens WHERE id = $1", oldID)
	if err != nil {
		r.logger.Error("error deleting old token", zap.Error(err))
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO tokens (id, jti, user_id, token, expires_at) VALUES ($1, $2, $3, $4, $5)", id, jti, userID, token, expiresAt)

	if err != nil {
		if database.IsPGError(err, database.PGUniqueViolationCode) {
			return domain.ErrTokenExists
		}
		r.logger.Error("error creating token", zap.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("error committing transaction", zap.Error(err))
		return err
	}

	return nil
}

// NewPostgresqlTokenRepo - конструктор для создания нового экземпляра PostgresqlTokenRepo.
func NewPostgresqlTokenRepo(db *sqlx.DB, logger *zap.Logger) repository.ITokenRepo {
	return &PostgresqlTokenRepo{
		db:     db,
		logger: logger,
	}
}
