package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

type DatabaseConfig struct {
	Host         string `env:"DB_HOST" env-default:"localhost"`
	Port         string `env:"DB_PORT" env-default:"5432"`
	User         string `env:"DB_USER" env-default:"postgres"`
	Pass         string `env:"DB_PASS" env-default:"postgres"`
	Name         string `env:"DB_NAME" env-default:"postgres"`
	SSLMode      string `env:"DB_SSLMODE" env-default:"disable"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" env-default:"5"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Pass, c.Name, c.SSLMode,
	)
}

type AuthConfig struct {
	JWTSecret  string `env:"JWT_SECRET" env-default:"secret"`
	AccessTTL  int    `env:"ACCESS_EXPIRATION_SECONDS" env-default:"1800"`    // Время жизни Access-токена в секундах, по умолчанию 30 минут
	RefreshTTL int    `env:"REFRESH_EXPIRATION_SECONDS" env-default:"604800"` // Время жизни Refresh-токена в секундах, по умолчанию 7 дней
}

type HTTPConfig struct {
	Host string `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port string `env:"HTTP_PORT" env-default:"8080"`
}

func (c HTTPConfig) Addr() string {
	if c.Host == "0.0.0.0" || c.Host == "localhost" {
		return fmt.Sprintf(":%s", c.Port)
	}
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type LoggerConfig struct {
	Level string `env:"LOG_LEVEL" env-default:"debug"` // debug, info, warn, error или silent
}

type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
	HTTP     HTTPConfig
	Logger   LoggerConfig
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
