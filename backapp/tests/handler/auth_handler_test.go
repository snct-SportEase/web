package handler_test

import (
	"backapp/internal/config"
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_GoogleLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("redirects LINE in-app browser users back to frontend", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{
			FrontendURL: "http://localhost:5173",
		}, mockUserRepo, mockEventRepo, mockClassRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/google/login", nil)
		c.Request.Header.Set("User-Agent", "Mozilla/5.0 Line/14.1.0")

		authHandler.GoogleLogin(c)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "http://localhost:5173/?error=line_inapp_browser_unsupported", w.Header().Get("Location"))
	})

	t.Run("sets oauth state cookie without secure attribute for local http", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/google/login", nil)

		authHandler.GoogleLogin(c)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

		var oauthStateCookie *http.Cookie
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == "oauthstate" {
				oauthStateCookie = cookie
				break
			}
		}

		if assert.NotNil(t, oauthStateCookie) {
			assert.False(t, oauthStateCookie.Secure)
			assert.True(t, oauthStateCookie.HttpOnly)
			assert.Equal(t, http.SameSiteLaxMode, oauthStateCookie.SameSite)
		}
	})

	t.Run("sets oauth state cookie with secure attribute for forwarded https", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/google/login", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")

		authHandler.GoogleLogin(c)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

		var oauthStateCookie *http.Cookie
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == "oauthstate" {
				oauthStateCookie = cookie
				break
			}
		}

		if assert.NotNil(t, oauthStateCookie) {
			assert.True(t, oauthStateCookie.Secure)
			assert.True(t, oauthStateCookie.HttpOnly)
			assert.Equal(t, http.SameSiteLaxMode, oauthStateCookie.SameSite)
		}
	})

	t.Run("redirects invalid callback state to frontend and clears state cookie", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{
			FrontendURL: "http://localhost:3300",
		}, mockUserRepo, mockEventRepo, mockClassRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/google/callback?state=query-state", nil)
		c.Request.AddCookie(&http.Cookie{Name: "oauthstate", Value: "cookie-state"})

		authHandler.GoogleCallback(c)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "http://localhost:3300/?error=invalid_state", w.Header().Get("Location"))

		var clearedCookie *http.Cookie
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == "oauthstate" {
				clearedCookie = cookie
				break
			}
		}

		if assert.NotNil(t, clearedCookie) {
			assert.Equal(t, -1, clearedCookie.MaxAge)
			assert.Equal(t, "/", clearedCookie.Path)
			assert.True(t, clearedCookie.HttpOnly)
			assert.Equal(t, http.SameSiteLaxMode, clearedCookie.SameSite)
		}
	})

	t.Run("redirects google callback errors to frontend", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{
			FrontendURL: "http://localhost:3300",
		}, mockUserRepo, mockEventRepo, mockClassRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/google/callback?error=access_denied", nil)

		authHandler.GoogleCallback(c)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "http://localhost:3300/?error=access_denied", w.Header().Get("Location"))
	})
}

func TestAuthHandler_UpdateUserClassRepByRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("replaces class role", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

		activeEventID := 3
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", 12).Return(&models.Class{
			ID:      12,
			EventID: &activeEventID,
			Name:    "2A",
		}, nil).Once()
		mockUserRepo.On("ReplaceClassRepRole", "user-1", "2A_rep", 12, &activeEventID).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/users/class-rep", bytes.NewBufferString(`{"user_id":"user-1","class_id":12}`))
		c.Request.Header.Set("Content-Type", "application/json")

		authHandler.UpdateUserClassRepByRoot(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Class role replaced successfully"}`, w.Body.String())
		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("rejects class outside active event", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

		activeEventID := 3
		otherEventID := 4
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", 12).Return(&models.Class{
			ID:      12,
			EventID: &otherEventID,
			Name:    "2A",
		}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/users/class-rep", bytes.NewBufferString(`{"user_id":"user-1","class_id":12}`))
		c.Request.Header.Set("Content-Type", "application/json")

		authHandler.UpdateUserClassRepByRoot(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Selected class does not belong to the active event"}`, w.Body.String())
		mockUserRepo.AssertNotCalled(t, "ReplaceClassRepRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestAuthHandler_PromoteUserByRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("replaces master role", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

		mockUserRepo.On("ReplaceMasterRole", "user-1", "student").Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/users/promote", bytes.NewBufferString(`{"user_id":"user-1","role":"student"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		authHandler.PromoteUserByRoot(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Master role replaced successfully"}`, w.Body.String())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("rejects non-master role", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/users/promote", bytes.NewBufferString(`{"user_id":"user-1","role":"judge"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		authHandler.PromoteUserByRoot(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"設定できるマスタロールは 'student' / 'admin' / 'root' のみです"}`, w.Body.String())
		mockUserRepo.AssertNotCalled(t, "ReplaceMasterRole", mock.Anything, mock.Anything)
	})
}

func TestAuthHandler_DemoteUserByRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUserRepo := new(MockUserRepository)
	mockEventRepo := new(MockEventRepository)
	mockClassRepo := new(MockClassRepository)
	authHandler := handler.NewAuthHandler(&config.Config{}, mockUserRepo, mockEventRepo, mockClassRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/users/promote", bytes.NewBufferString(`{"user_id":"user-1","role":"admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	authHandler.DemoteUserByRoot(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"マスタロールは剥奪ではなく交換してください"}`, w.Body.String())
	mockUserRepo.AssertNotCalled(t, "DeleteUserRole", mock.Anything, mock.Anything)
}
