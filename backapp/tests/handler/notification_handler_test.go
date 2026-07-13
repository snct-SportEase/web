package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"backapp/internal/push"
	"bytes"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	mockEventRepo.On("GetActiveEvent").Return(3, nil).Once()
	mockRoleRepo.On("GetAllRoles").Return([]models.Role{
		{ID: 1, Name: "student"},
		{ID: 2, Name: "admin"},
	}, nil).Once()
	mockNotifRepo.On("CreateNotification", "大会のお知らせ", "本文です", "general", "user-1", mock.MatchedBy(func(eventID *int) bool {
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
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

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
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

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
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}
	authKey, p256dhKey := validPushSubscriptionKeys(t)
	endpoint := "https://fcm.googleapis.com/fcm/send/test-token"

	// Save
	{
		mockNotifRepo.On("UpsertPushSubscription", "user-1", endpoint, authKey, p256dhKey, push.MaxSubscriptionsPerUser).
			Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		subReq := map[string]any{
			"endpoint": endpoint,
			"keys": map[string]string{
				"auth":   authKey,
				"p256dh": p256dhKey,
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
		mockNotifRepo.On("DeletePushSubscription", "user-1", endpoint).
			Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		payload, _ := json.Marshal(map[string]string{"endpoint": endpoint})
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/notifications/subscription", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user", user)

		h.DeleteSubscription(c)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	mockNotifRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func validPushSubscriptionKeys(t *testing.T) (string, string) {
	t.Helper()
	auth := []byte("0123456789abcdef")
	_, x, y, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate push subscription key: %v", err)
	}
	publicKey := elliptic.Marshal(elliptic.P256(), x, y)
	return base64.RawURLEncoding.EncodeToString(auth), base64.RawURLEncoding.EncodeToString(publicKey)
}

func TestNotificationHandler_SaveSubscriptionRejectsUnsafeEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockNotifRepo := new(MockNotificationRepository)
	h := handler.NewNotificationHandler(mockNotifRepo, nil, nil, nil, "", "")
	authKey, p256dhKey := validPushSubscriptionKeys(t)

	payload, _ := json.Marshal(map[string]any{
		"endpoint": "https://127.0.0.1/internal",
		"keys": map[string]string{
			"auth":   authKey,
			"p256dh": p256dhKey,
		},
	})
	request := httptest.NewRequest(http.MethodPost, "/api/notifications/subscription", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	context.Set("user", &models.User{ID: "user-1"})

	h.SaveSubscription(context)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	mockNotifRepo.AssertNumberOfCalls(t, "UpsertPushSubscription", 0)
}

func TestNotificationHandler_SaveSubscriptionRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockNotifRepo := new(MockNotificationRepository)
	h := handler.NewNotificationHandler(mockNotifRepo, nil, nil, nil, "", "")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/notifications/subscription",
		strings.NewReader(`{"endpoint":"https://fcm.googleapis.com/`+strings.Repeat("a", 5000)+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	context.Set("user", &models.User{ID: "user-1"})

	h.SaveSubscription(context)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	mockNotifRepo.AssertNumberOfCalls(t, "UpsertPushSubscription", 0)
}

func TestNotificationHandler_GetSubscription_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	expectedSubs := []models.PushSubscription{
		{
			ID:        1,
			UserID:    "user-1",
			Endpoint:  "https://example.com/push/1",
			AuthKey:   "auth-key-1",
			P256dhKey: "p256dh-key-1",
		},
		{
			ID:        2,
			UserID:    "user-1",
			Endpoint:  "https://example.com/push/2",
			AuthKey:   "auth-key-2",
			P256dhKey: "p256dh-key-2",
		},
	}

	mockNotifRepo.On("GetPushSubscriptionsByUserID", "user-1").
		Return(expectedSubs, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/api/notifications/subscription", nil)
	c.Set("user", user)

	h.GetSubscription(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["subscribed"].(bool))
	assert.Equal(t, float64(2), response["count"].(float64))

	endpoints, ok := response["endpoints"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, endpoints, 2)
	assert.Contains(t, endpoints, "https://example.com/push/1")
	assert.Contains(t, endpoints, "https://example.com/push/2")

	mockNotifRepo.AssertExpectations(t)
}

func TestNotificationHandler_GetSubscription_NoSubscriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	mockNotifRepo.On("GetPushSubscriptionsByUserID", "user-1").
		Return([]models.PushSubscription{}, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/api/notifications/subscription", nil)
	c.Set("user", user)

	h.GetSubscription(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["subscribed"].(bool))
	assert.Equal(t, float64(0), response["count"].(float64))

	endpoints, ok := response["endpoints"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, endpoints, 0)

	mockNotifRepo.AssertExpectations(t)
}

func TestNotificationHandler_GetSubscription_UnauthorizedWithoutUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/api/notifications/subscription", nil)

	h.GetSubscription(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockNotifRepo.AssertNotCalled(t, "GetPushSubscriptionsByUserID", mock.Anything)
}

func TestNotificationHandler_GetSubscription_RepositoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	mockNotifRepo.On("GetPushSubscriptionsByUserID", "user-1").
		Return(nil, assert.AnError).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/api/notifications/subscription", nil)
	c.Set("user", user)

	h.GetSubscription(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "購読情報の取得に失敗しました")

	mockNotifRepo.AssertExpectations(t)
}

func TestNotificationHandler_GetPushSubscriptionStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	mockRoleRepo.On("GetAllRoles").Return([]models.Role{
		{ID: 1, Name: "student"},
		{ID: 2, Name: "admin"},
	}, nil).Once()
	mockNotifRepo.On("GetPushSubscriptionStatsByRoles", []string{"admin", "student"}).
		Return(models.PushSubscriptionStats{
			TargetUserCount:           10,
			SubscribedUserCount:       6,
			SubscriptionEndpointCount: 8,
		}, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/api/root/notifications/subscription-stats?roles=Student&roles=admin", nil)

	h.GetPushSubscriptionStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"admin", "student"}, response["target_roles"])

	stats := response["stats"].(map[string]interface{})
	assert.Equal(t, float64(10), stats["target_user_count"])
	assert.Equal(t, float64(6), stats["subscribed_user_count"])
	assert.Equal(t, float64(8), stats["subscription_endpoint_count"])

	mockRoleRepo.AssertExpectations(t)
	mockNotifRepo.AssertExpectations(t)
}

func TestNotificationHandler_UpdateNotificationFilters_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	reqBody := map[string]interface{}{
		"filters": []string{"general", "match_my_class", "finals"},
	}
	payload, _ := json.Marshal(reqBody)

	mockUserRepo.On("UpdateNotificationFilters", "user-1", []string{"general", "match_my_class", "finals"}).Return(nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPut, "/api/user/notification-filters", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	h.UpdateNotificationFilters(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "通知フィルタを更新しました", response["message"])
	assert.Equal(t, []interface{}{"general", "match_my_class", "finals"}, response["filters"])

	mockUserRepo.AssertExpectations(t)
}

func TestNotificationHandler_UpdateNotificationFilters_AddGeneral(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	reqBody := map[string]interface{}{
		"filters": []string{"match_my_class", "finals"},
	}
	payload, _ := json.Marshal(reqBody)

	mockUserRepo.On("UpdateNotificationFilters", "user-1", []string{"match_my_class", "finals", "general"}).Return(nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPut, "/api/user/notification-filters", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	h.UpdateNotificationFilters(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "通知フィルタを更新しました", response["message"])
	assert.Equal(t, []interface{}{"match_my_class", "finals", "general"}, response["filters"])

	mockUserRepo.AssertExpectations(t)
}

func TestNotificationHandler_UpdateNotificationFilters_InvalidFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	reqBody := map[string]interface{}{
		"filters": []string{"general", "invalid_filter"},
	}
	payload, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPut, "/api/user/notification-filters", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	h.UpdateNotificationFilters(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "無効なフィルタが含まれています")

	mockUserRepo.AssertNotCalled(t, "UpdateNotificationFilters", mock.Anything, mock.Anything)
}

func TestNotificationHandler_UpdateNotificationFilters_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	reqBody := map[string]interface{}{
		"filters": []string{"general", "match_my_class"},
	}
	payload, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPut, "/api/user/notification-filters", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateNotificationFilters(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ユーザー情報を取得できませんでした", response["error"])

	mockUserRepo.AssertNotCalled(t, "UpdateNotificationFilters", mock.Anything, mock.Anything)
}

func TestNotificationHandler_UpdateNotificationFilters_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPut, "/api/user/notification-filters", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	h.UpdateNotificationFilters(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "不正なリクエスト形式です", response["error"])

	mockUserRepo.AssertNotCalled(t, "UpdateNotificationFilters", mock.Anything, mock.Anything)
}

func TestNotificationHandler_UpdateNotificationFilters_RepositoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	user := &models.User{ID: "user-1"}

	reqBody := map[string]interface{}{
		"filters": []string{"general", "all_matches"},
	}
	payload, _ := json.Marshal(reqBody)

	mockUserRepo.On("UpdateNotificationFilters", "user-1", []string{"general", "all_matches"}).Return(assert.AnError).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPut, "/api/user/notification-filters", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	h.UpdateNotificationFilters(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "通知フィルタの更新に失敗しました", response["error"])

	mockUserRepo.AssertExpectations(t)
}

func TestNotificationHandler_CreateNotification_WithType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	mockEventRepo.On("GetActiveEvent").Return(3, nil).Once()
	mockRoleRepo.On("GetAllRoles").Return([]models.Role{
		{ID: 1, Name: "student"},
		{ID: 2, Name: "admin"},
	}, nil).Once()
	mockNotifRepo.On("CreateNotification", "決勝戦のお知らせ", "決勝戦が始まります", "finals", "user-1", mock.MatchedBy(func(eventID *int) bool {
		return eventID != nil && *eventID == 3
	})).Return(int64(11), nil).Once()
	mockNotifRepo.On("AddNotificationTargets", int64(11), []string{"admin", "student"}).Return(nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]any{
		"title":        "決勝戦のお知らせ",
		"body":         "決勝戦が始まります",
		"type":         "finals",
		"target_roles": []string{"student", "admin"},
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

func TestNotificationHandler_CreateNotification_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNotifRepo := new(MockNotificationRepository)
	mockEventRepo := new(MockEventRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockUserRepo := new(MockUserRepository)

	h := handler.NewNotificationHandler(mockNotifRepo, mockEventRepo, mockRoleRepo, mockUserRepo, "", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]any{
		"title":        "お知らせ",
		"body":         "本文です",
		"type":         "invalid_type",
		"target_roles": []string{"student"},
	}
	payload, _ := json.Marshal(reqBody)

	c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/notifications", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", &models.User{ID: "user-1"})

	h.CreateNotification(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "無効な通知タイプです")

	mockNotifRepo.AssertNotCalled(t, "CreateNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
