package handler_test

import (
	"backapp/internal/config"
	"backapp/internal/handler"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
