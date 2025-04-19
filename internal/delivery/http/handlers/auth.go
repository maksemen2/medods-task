package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maksemen2/medods-task/internal/delivery/http/dto"
	"github.com/maksemen2/medods-task/internal/domain"
	"github.com/maksemen2/medods-task/internal/service"
	"go.uber.org/zap"
	"net/http"
)

// AuthHandler - структура для обработки запросов аутентификации.
type AuthHandler struct {
	logger  *zap.Logger
	service service.IAuthService
}

func NewAuthHandler(logger *zap.Logger, service service.IAuthService) *AuthHandler {
	return &AuthHandler{
		logger:  logger,
		service: service,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/auth", h.GETAuth)
	router.POST("/refresh", h.POSTRefresh)
}

// handleError - хелпер для обработки доменных ошибок, возвращаемых сервисом
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUnexpected):
		c.AbortWithStatus(http.StatusInternalServerError)
	case errors.Is(err, domain.ErrInvalidAccessToken), errors.Is(err, domain.ErrInvalidRefreshToken), errors.Is(err, domain.ErrTokenNotFound):
		c.AbortWithStatus(http.StatusBadRequest)
	default:
		h.logger.Error("unexpected error from authService", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (h *AuthHandler) GETAuth(c *gin.Context) {
	var query dto.AuthQueryParams

	if err := c.ShouldBindQuery(&query); err != nil {
		h.logger.Debug("error binding query", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	guid, err := uuid.Parse(query.GUID)

	if err != nil {
		h.logger.Debug("error parsing guid", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	domainAuth, err := h.service.AuthenticateUser(c.Request.Context(), guid, c.ClientIP())

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  domainAuth.AccessToken,
		RefreshToken: domainAuth.RefreshToken,
	})
}

func (h *AuthHandler) POSTRefresh(c *gin.Context) {
	var req dto.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("error binding json", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	domainAuth, err := h.service.RefreshToken(c.Request.Context(), req.AccessToken, req.RefreshToken, c.ClientIP())

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  domainAuth.AccessToken,
		RefreshToken: domainAuth.RefreshToken,
	})
}
