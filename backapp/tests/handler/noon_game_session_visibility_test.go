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

func noonGameStudentContext(c *gin.Context) {
	c.Set("user", &models.User{Roles: []models.Role{{Name: "student"}}})
}

func TestNoonGameHandler_ListSessions_HidesDraftsFromStudents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noonRepo := new(MockNoonGameRepository)
	classRepo := new(MockClassRepository)
	eventRepo := new(MockEventRepository)
	h := handler.NewNoonGameHandler(noonRepo, classRepo, eventRepo)

	eventRepo.On("GetEventByID", 1).Return(&models.Event{ID: 1}, nil).Once()
	noonRepo.On("ListSessionsByEvent", 1, true).Return([]*models.NoonGameSession{{ID: 2, EventID: 1, Name: "綱引き", Status: "published"}}, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "event_id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/student/events/1/noon-game/sessions", nil)
	noonGameStudentContext(c)
	h.ListSessions(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response struct {
		Sessions []*models.NoonGameSession `json:"sessions"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Len(t, response.Sessions, 1)
	assert.Equal(t, "published", response.Sessions[0].Status)
	noonRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestNoonGameHandler_GetSessionByID_HidesDraftFromStudents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noonRepo := new(MockNoonGameRepository)
	h := handler.NewNoonGameHandler(noonRepo, new(MockClassRepository), new(MockEventRepository))
	noonRepo.On("GetSessionByID", 10).Return(&models.NoonGameSession{ID: 10, EventID: 1, Status: "draft"}, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "session_id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/student/events/1/noon-game/sessions/10", nil)
	noonGameStudentContext(c)
	h.GetSessionByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	noonRepo.AssertExpectations(t)
}

func TestNoonGameHandler_UpsertSession_DefaultsToDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noonRepo := new(MockNoonGameRepository)
	classRepo := new(MockClassRepository)
	eventRepo := new(MockEventRepository)
	h := handler.NewNoonGameHandler(noonRepo, classRepo, eventRepo)

	eventRepo.On("GetEventByID", 1).Return(&models.Event{ID: 1}, nil).Once()
	noonRepo.On("UpsertSession", mock.MatchedBy(func(session *models.NoonGameSession) bool {
		return session.EventID == 1 && session.Status == "draft" && session.TemplateKey == "custom"
	})).Return(&models.NoonGameSession{ID: 10, EventID: 1, Name: "綱引き", Mode: "mixed", Status: "draft", TemplateKey: "custom"}, nil).Once()
	classRepo.On("GetAllClasses", 1).Return([]*models.Class{}, nil).Once()
	noonRepo.On("GetGroupsWithMembers", 10).Return([]*models.NoonGameGroupWithMembers{}, nil).Once()
	noonRepo.On("GetMatchesWithResults", 10).Return([]*models.NoonGameMatchWithResult{}, nil).Once()
	noonRepo.On("SumPointsByClass", 10).Return(map[int]int{}, nil).Once()
	noonRepo.On("ListTemplateRunsBySession", 10).Return([]*models.NoonGameTemplateRun{}, nil).Once()

	body := []byte(`{"name":"綱引き","mode":"mixed"}`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "event_id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/root/events/1/noon-game/sessions", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.UpsertSession(c)

	assert.Equal(t, http.StatusOK, w.Code)
	noonRepo.AssertExpectations(t)
	classRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}
