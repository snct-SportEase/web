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

type MockNotificationRequestRepository struct {
	mock.Mock
}

func (m *MockNotificationRequestRepository) CreateRequest(request *models.NotificationRequest) (int64, error) {
	args := m.Called(request)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRequestRepository) UpdateRequestStatus(id int, status models.NotificationRequestStatus, resolverID *string) error {
	return m.Called(id, status, resolverID).Error(0)
}

func (m *MockNotificationRequestRepository) GetRequestByID(id int) (*models.NotificationRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationRequest), args.Error(1)
}

func (m *MockNotificationRequestRepository) GetRequestsByRequester(requesterID string) ([]*models.NotificationRequest, error) {
	args := m.Called(requesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NotificationRequest), args.Error(1)
}

func (m *MockNotificationRequestRepository) GetAllRequests() ([]*models.NotificationRequest, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NotificationRequest), args.Error(1)
}

func (m *MockNotificationRequestRepository) AddMessage(requestID int, senderID, message string) (int64, error) {
	args := m.Called(requestID, senderID, message)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRequestRepository) GetMessages(requestID int) ([]*models.NotificationRequestMessage, error) {
	args := m.Called(requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NotificationRequestMessage), args.Error(1)
}

func (m *MockNotificationRequestRepository) GetParticipants(requestID int) (string, *string, error) {
	args := m.Called(requestID)
	var resolverID *string
	if args.Get(1) != nil {
		resolverID = args.Get(1).(*string)
	}
	return args.String(0), resolverID, args.Error(2)
}

func newNotificationRequestHandler(requestRepo *MockNotificationRequestRepository) *handler.NotificationRequestHandler {
	return handler.NewNotificationRequestHandler(
		requestRepo,
		new(MockNotificationRepository),
		new(MockRoleRepository),
		"",
		"",
	)
}

func notificationRequestContext(method, target string, body any, user *models.User) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		requestBody = bytes.NewReader(payload)
	}
	c.Request, _ = http.NewRequest(method, target, requestBody)
	c.Request.Header.Set("Content-Type", "application/json")
	if user != nil {
		c.Set("user", user)
	}
	return c, w
}

func TestNotificationRequestHandler_CreateRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(MockNotificationRequestRepository)
	h := newNotificationRequestHandler(repo)
	user := &models.User{ID: "student-1", Roles: []models.Role{{Name: "student"}}}

	repo.On("CreateRequest", mock.MatchedBy(func(request *models.NotificationRequest) bool {
		return request.Title == "決勝戦の招集" &&
			request.Body == "10分前に集合してください" &&
			request.TargetText == "決勝進出クラス" &&
			request.RequesterID == user.ID &&
			request.Status == models.NotificationRequestStatusPending
	})).Return(int64(42), nil).Once()

	c, w := notificationRequestContext(http.MethodPost, "/api/student/notification-requests", map[string]string{
		"title":       "  決勝戦の招集  ",
		"body":        "  10分前に集合してください  ",
		"target_text": "  決勝進出クラス  ",
	}, user)
	h.CreateRequest(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"request_id":42}`, w.Body.String())
	repo.AssertExpectations(t)
}

func TestNotificationRequestHandler_ListRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(MockNotificationRequestRepository)
	h := newNotificationRequestHandler(repo)
	student := &models.User{ID: "student-1", Roles: []models.Role{{Name: "student"}}}
	requests := []*models.NotificationRequest{{ID: 1, Title: "集合連絡", RequesterID: student.ID}}

	repo.On("GetRequestsByRequester", student.ID).Return(requests, nil).Once()
	studentContext, studentResponse := notificationRequestContext(http.MethodGet, "/api/student/notification-requests", nil, student)
	h.ListStudentRequests(studentContext)
	assert.Equal(t, http.StatusOK, studentResponse.Code)
	assert.Contains(t, studentResponse.Body.String(), "集合連絡")

	repo.On("GetAllRequests").Return(requests, nil).Once()
	rootContext, rootResponse := notificationRequestContext(http.MethodGet, "/api/root/notification-requests", nil, student)
	h.ListRootRequests(rootContext)
	assert.Equal(t, http.StatusOK, rootResponse.Code)
	repo.AssertExpectations(t)
}

func TestNotificationRequestHandler_GetRequestDetailEnforcesOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(MockNotificationRequestRepository)
	h := newNotificationRequestHandler(repo)
	request := &models.NotificationRequest{ID: 7, Title: "集合連絡", RequesterID: "student-1"}
	repo.On("GetRequestByID", 7).Return(request, nil).Once()

	otherStudent := &models.User{ID: "student-2", Roles: []models.Role{{Name: "student"}}}
	c, w := notificationRequestContext(http.MethodGet, "/api/student/notification-requests/7", nil, otherStudent)
	c.Params = gin.Params{{Key: "request_id", Value: "7"}}
	h.GetRequestDetail(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "アクセス権限")
	repo.AssertNotCalled(t, "GetMessages", mock.Anything)
	repo.AssertExpectations(t)
}

func TestNotificationRequestHandler_AddMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(MockNotificationRequestRepository)
	h := newNotificationRequestHandler(repo)
	student := &models.User{ID: "student-1", Roles: []models.Role{{Name: "student"}}}
	request := &models.NotificationRequest{ID: 7, RequesterID: student.ID}
	repo.On("GetRequestByID", 7).Return(request, nil).Once()
	repo.On("AddMessage", 7, student.ID, "集合場所は第1体育館です").Return(int64(9), nil).Once()

	c, w := notificationRequestContext(http.MethodPost, "/api/student/notification-requests/7/messages", map[string]string{
		"message": "  集合場所は第1体育館です  ",
	}, student)
	c.Params = gin.Params{{Key: "request_id", Value: "7"}}
	h.AddMessage(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}

func TestNotificationRequestHandler_DecideRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(MockNotificationRequestRepository)
	h := newNotificationRequestHandler(repo)
	root := &models.User{ID: "root-1", Roles: []models.Role{{Name: "root"}}}
	request := &models.NotificationRequest{ID: 7, RequesterID: "student-1"}
	repo.On("GetRequestByID", 7).Return(request, nil).Once()
	repo.On("UpdateRequestStatus", 7, models.NotificationRequestStatusApproved, mock.MatchedBy(func(id *string) bool {
		return id != nil && *id == root.ID
	})).Return(nil).Once()

	c, w := notificationRequestContext(http.MethodPost, "/api/root/notification-requests/7/decision", map[string]string{
		"status": "approved",
	}, root)
	c.Params = gin.Params{{Key: "request_id", Value: "7"}}
	h.DecideRequest(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}
