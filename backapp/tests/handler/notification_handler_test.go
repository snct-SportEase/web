package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationHandler_CreateNotification_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, "", "")

	mockEventRepo.On("GetActiveEvent").Return(3, nil).Once()
	mockRoleRepo.On("GetAllRoles").Return([]models.Role{
		{ID: 1, Name: "student"},
		{ID: 2, Name: "admin"},
	}, nil).Once()
	mockNotifRepo.On("CreateNotification", "大会のお知らせ", "本文です", "user-1", mock.MatchedBy(func(eventID *int) bool {
		return eventID != nil && *eventID == 3
	})).Return(int64(10), nil).Once()
	mockNotifRepo.On("AddNotificationTargets", int64(10), []string{"admin", "student"}).Return(nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]any{
		"title":        "大会のお知らせ",
		"body":         "本文です",
		"target_roles": []string{"Student", "admin"},
	}
	payload, _ := json.Marshal(reqBody)

	c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/notifications", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", &models.User{ID: "user-1"})

	h.CreateNotification(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockNotifRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestNotificationHandler_CreateNotification_UnauthorizedWithoutUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, "", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]any{
		"title":        "大会のお知らせ",
		"body":         "本文です",
		"target_roles": []string{"student"},
	}
	payload, _ := json.Marshal(reqBody)

	c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/notifications", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateNotification(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockNotifRepo.AssertNotCalled(t, "CreateNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockRoleRepo.AssertExpectations(t)
}

func TestNotificationHandler_ListNotifications_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, "", "")

	expected := []models.Notification{
		{ID: 1, Title: "お知らせ", Body: "内容"},
	}

	mockNotifRepo.On("GetNotificationsForAccess", []string{"student"}, "user-1", true, 10).
		Return(expected, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/api/notifications?include_authored=true&limit=10", nil)
	c.Set("user", &models.User{
		ID: "user-1",
		Roles: []models.Role{
			{Name: "student"},
		},
	})

	h.ListNotifications(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockNotifRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestNotificationHandler_SaveAndDeleteSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, "", "")

	user := &models.User{ID: "user-1"}

	// Save
	{
		mockNotifRepo.On("UpsertPushSubscription", "user-1", "https://example.com", "auth", "p256dh").
			Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		subReq := map[string]any{
			"endpoint": "https://example.com",
			"keys": map[string]string{
				"auth":   "auth",
				"p256dh": "p256dh",
			},
		}
		payload, _ := json.Marshal(subReq)

		c.Request, _ = http.NewRequest(http.MethodPost, "/api/notifications/subscription", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user", user)

		h.SaveSubscription(c)

		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Delete
	{
		mockNotifRepo.On("DeletePushSubscription", "user-1", "https://example.com").
			Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		payload, _ := json.Marshal(map[string]string{"endpoint": "https://example.com"})
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/notifications/subscription", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user", user)

		h.DeleteSubscription(c)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	mockNotifRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
