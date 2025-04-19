package database

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// TxRollback - хелпер для отката транзакции.
// Предполагается, что эта функция должна использоваться в defer
// Если транзакция уже завершена, то ошибка будет игнорироваться
func TxRollback(tx *sqlx.Tx, logger *zap.Logger) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		logger.Error("failed to rollback transaction", zap.Error(err))
	}
}
