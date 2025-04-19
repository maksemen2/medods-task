package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/medods-task/internal/delivery/http/handlers"
	"github.com/maksemen2/medods-task/internal/pkg/log"
	"github.com/maksemen2/medods-task/internal/service"
	"go.uber.org/zap"
)

// New настраивает роутинг приложения и устанавливает мидлвари.
// Возвращает инстанс gin.Engine
func New(logger *zap.Logger, authService service.IAuthService) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), log.NewMiddleware(logger))

	authGroup := router.Group("")

	authHandler := handlers.NewAuthHandler(logger, authService)

	authHandler.RegisterRoutes(authGroup)

	return router
}
