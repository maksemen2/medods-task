package log

import (
	"fmt"
	"go.uber.org/zap"
)

// NewZapLogger создает новый экземпляр логгера с использованием конфигурации.
// Поле level должно быть одним из: "debug", "info", "warn", "error" или "silent".
// При уровне "silent" возвращается пустой логгер (zap.NewNop).
func NewZapLogger(level string) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()

	switch level {
	case "debug":
		zapConfig.Level.SetLevel(zap.DebugLevel)
	case "info":
		zapConfig.Level.SetLevel(zap.InfoLevel)
	case "warn":
		zapConfig.Level.SetLevel(zap.WarnLevel)
	case "error":
		zapConfig.Level.SetLevel(zap.ErrorLevel)
	case "silent":
		return zap.NewNop(), nil
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	return zapConfig.Build()
}
