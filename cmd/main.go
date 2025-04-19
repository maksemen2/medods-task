package main

import (
	"context"
	"errors"
	"github.com/maksemen2/medods-task/internal/config"
	"github.com/maksemen2/medods-task/internal/delivery/http/routes"
	"github.com/maksemen2/medods-task/internal/pkg/auth/jwt"
	"github.com/maksemen2/medods-task/internal/pkg/database"
	"github.com/maksemen2/medods-task/internal/pkg/log"
	postgresqlrepo "github.com/maksemen2/medods-task/internal/repository/postgresql"
	"github.com/maksemen2/medods-task/internal/service"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	logger, err := log.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}

	db, err := database.NewPostgresDB(cfg.Database.DSN(), cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	tokenRepo := postgresqlrepo.NewPostgresqlTokenRepo(db, logger)
	userRepo := postgresqlrepo.NewPostgresqlUserRepo(db, logger)
	tokenManager := jwt.NewManager([]byte(cfg.Auth.JWTSecret), time.Duration(cfg.Auth.AccessTTL)*time.Second)

	router := routes.New(logger, service.NewAuthServiceImpl(userRepo, tokenRepo, tokenManager, logger, time.Duration(cfg.Auth.RefreshTTL)*time.Second))

	srv := &http.Server{
		Addr:    cfg.HTTP.Addr(),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Server started", zap.String("addr", srv.Addr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Server exiting")
}
