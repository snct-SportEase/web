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

func TestNoonGameHandler_UpsertSession_SyncsNoonGameSport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNoonRepo := new(MockNoonGameRepository)
	mockClassRepo := new(MockClassRepository)
	mockEventRepo := new(MockEventRepository)
	mockSportRepo := new(MockSportRepository)

	h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo).WithSportSync(mockSportRepo)

	eventID := 1
	sessionID := 10
	sessionName := "綱引き"

	mockEventRepo.On("GetEventByID", eventID).Return(&models.Event{ID: eventID, Name: "2025春季大会"}, nil).Once()
	mockNoonRepo.
		On("UpsertSession", mock.MatchedBy(func(session *models.NoonGameSession) bool {
			return session != nil && session.EventID == eventID && session.Name == sessionName && session.Mode == "mixed"
		})).
		Return(&models.NoonGameSession{
			ID:      sessionID,
			EventID: eventID,
			Name:    sessionName,
			Mode:    "mixed",
		}, nil).
		Once()

	mockSportRepo.On("GetSportByName", sessionName).Return(nil, nil).Once()
	mockSportRepo.
		On("CreateSport", mock.MatchedBy(func(sport *models.Sport) bool {
			return sport != nil && sport.Name == sessionName
		})).
		Return(int64(7), nil).
		Once()
	mockSportRepo.On("GetSportsByEventID", eventID).Return([]*models.EventSport{}, nil).Once()
	mockSportRepo.
		On("AssignSportToEvent", mock.MatchedBy(func(eventSport *models.EventSport) bool {
			return eventSport != nil &&
				eventSport.EventID == eventID &&
				eventSport.SportID == 7 &&
				eventSport.Location == "noon_game" &&
				eventSport.RulesType == "markdown"
		})).
		Return(nil).
		Once()

	mockClassRepo.On("GetAllClasses", eventID).Return([]*models.Class{{ID: 101, Name: "1A"}}, nil).Once()
	mockNoonRepo.On("GetGroupsWithMembers", sessionID).Return([]*models.NoonGameGroupWithMembers{}, nil).Once()
	mockNoonRepo.On("GetMatchesWithResults", sessionID).Return([]*models.NoonGameMatchWithResult{}, nil).Once()
	mockNoonRepo.On("SumPointsByClass", sessionID).Return(map[int]int{}, nil).Once()
	mockNoonRepo.On("ListTemplateRunsBySession", sessionID).Return([]*models.NoonGameTemplateRun{}, nil).Once()

	body, err := json.Marshal(map[string]interface{}{
		"name":                 sessionName,
		"mode":                 "mixed",
		"win_points":           0,
		"loss_points":          0,
		"draw_points":          0,
		"participation_points": 0,
		"allow_manual_points":  true,
		"description":          "",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/root/events/1/noon-game/session", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "event_id", Value: "1"}}
	c.Request = req

	h.UpsertSession(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockNoonRepo.AssertExpectations(t)
	mockClassRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
	mockSportRepo.AssertExpectations(t)
}

func TestNoonGameHandler_GetSession_SyncsExistingNoonGameSport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNoonRepo := new(MockNoonGameRepository)
	mockClassRepo := new(MockClassRepository)
	mockEventRepo := new(MockEventRepository)
	mockSportRepo := new(MockSportRepository)

	h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo).WithSportSync(mockSportRepo)

	eventID := 1
	sessionID := 10
	sessionName := "綱引き"
	session := &models.NoonGameSession{
		ID:      sessionID,
		EventID: eventID,
		Name:    sessionName,
		Mode:    "mixed",
	}

	mockEventRepo.On("GetEventByID", eventID).Return(&models.Event{ID: eventID, Name: "2025春季大会"}, nil).Once()
	mockNoonRepo.On("GetSessionByEvent", eventID).Return(session, nil).Once()

	mockSportRepo.On("GetSportByName", sessionName).Return(&models.Sport{ID: 7, Name: sessionName}, nil).Once()
	mockSportRepo.On("GetSportsByEventID", eventID).Return([]*models.EventSport{}, nil).Once()
	mockSportRepo.
		On("AssignSportToEvent", mock.MatchedBy(func(eventSport *models.EventSport) bool {
			return eventSport != nil &&
				eventSport.EventID == eventID &&
				eventSport.SportID == 7 &&
				eventSport.Location == "noon_game"
		})).
		Return(nil).
		Once()

	mockClassRepo.On("GetAllClasses", eventID).Return([]*models.Class{{ID: 101, Name: "1A"}}, nil).Once()
	mockNoonRepo.On("GetGroupsWithMembers", sessionID).Return([]*models.NoonGameGroupWithMembers{}, nil).Once()
	mockNoonRepo.On("GetMatchesWithResults", sessionID).Return([]*models.NoonGameMatchWithResult{}, nil).Once()
	mockNoonRepo.On("SumPointsByClass", sessionID).Return(map[int]int{}, nil).Once()
	mockNoonRepo.On("ListTemplateRunsBySession", sessionID).Return([]*models.NoonGameTemplateRun{}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/root/events/1/noon-game/session", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "event_id", Value: "1"}}
	c.Request = req

	h.GetSession(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockNoonRepo.AssertExpectations(t)
	mockClassRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
	mockSportRepo.AssertExpectations(t)
}
