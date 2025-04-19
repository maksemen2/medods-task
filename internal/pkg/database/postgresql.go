package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Коды ошибок PostgreSQL.
// См. https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	PGUniqueViolationCode = "23505" // Уникальное ограничение нарушено
)

func NewPostgresDB(dsn string, maxOpenConns, maxIdleConns int) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	return db, nil
}

// IsPGError - проверяет, является ли ошибка
// ошибкой PostgreSQL с указанным кодом.
func IsPGError(err error, code string) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == pq.ErrorCode(code)
	}

	return false
}
