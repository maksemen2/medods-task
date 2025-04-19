package handlers_test

import (
	"bytes"
	"encoding/json"
	"github.com/maksemen2/medods-task/internal/delivery/http/dto"
	"github.com/maksemen2/medods-task/internal/delivery/http/handlers"
	"github.com/maksemen2/medods-task/internal/domain"
	mock_service "github.com/maksemen2/medods-task/internal/service/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_GETAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockIAuthService(ctrl)
		logger := zap.NewNop()
		h := handlers.NewAuthHandler(logger, mockService)

		router := gin.New()
		router.GET("/auth", h.GETAuth)

		guid := uuid.New()
		expected := dto.AuthResponse{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}

		mockService.EXPECT().AuthenticateUser(gomock.Any(), guid, gomock.Any()).
			Return(&domain.UserAuth{
				AccessToken:  expected.AccessToken,
				RefreshToken: expected.RefreshToken,
			}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth?guid="+guid.String(), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.AuthResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expected, response)
	})

	t.Run("invalid guid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockIAuthService(ctrl)
		logger := zap.NewNop()
		h := handlers.NewAuthHandler(logger, mockService)

		router := gin.New()
		router.GET("/auth", h.GETAuth)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth?guid=invalid", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockIAuthService(ctrl)
		logger := zap.NewNop()
		h := handlers.NewAuthHandler(logger, mockService)

		router := gin.New()
		router.GET("/auth", h.GETAuth)

		guid := uuid.New()
		mockService.EXPECT().AuthenticateUser(gomock.Any(), guid, gomock.Any()).
			Return(nil, domain.ErrUnexpected)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth?guid="+guid.String(), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAuthHandler_POSTRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockIAuthService(ctrl)
		logger := zap.NewNop()
		h := handlers.NewAuthHandler(logger, mockService)

		router := gin.New()
		router.POST("/refresh", h.POSTRefresh)

		request := dto.RefreshRequest{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}
		expected := dto.AuthResponse{
			AccessToken:  "new_access",
			RefreshToken: "new_refresh",
		}

		mockService.EXPECT().RefreshToken(gomock.Any(), request.AccessToken, request.RefreshToken, gomock.Any()).
			Return(&domain.UserAuth{
				AccessToken:  expected.AccessToken,
				RefreshToken: expected.RefreshToken,
			}, nil)

		body, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.AuthResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expected, response)
	})

	t.Run("invalid request body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockIAuthService(ctrl)
		logger := zap.NewNop()
		h := handlers.NewAuthHandler(logger, mockService)

		router := gin.New()
		router.POST("/refresh", h.POSTRefresh)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader([]byte("{invalid}")))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockIAuthService(ctrl)
		logger := zap.NewNop()
		h := handlers.NewAuthHandler(logger, mockService)

		router := gin.New()
		router.POST("/refresh", h.POSTRefresh)

		request := dto.RefreshRequest{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}

		mockService.EXPECT().RefreshToken(gomock.Any(), request.AccessToken, request.RefreshToken, gomock.Any()).
			Return(nil, domain.ErrInvalidAccessToken)

		body, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
