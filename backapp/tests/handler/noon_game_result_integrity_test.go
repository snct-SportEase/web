package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"backapp/internal/repository"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecordMatchResultRejectsDuplicateRankingsBeforeWriting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noonRepo := new(MockNoonGameRepository)
	classRepo := new(MockClassRepository)
	eventRepo := new(MockEventRepository)
	h := handler.NewNoonGameHandler(noonRepo, classRepo, eventRepo)

	class1, class2 := 101, 102
	entries := []*models.NoonGameMatchEntry{
		{ID: 21, SideType: "class", ClassID: &class1},
		{ID: 22, SideType: "class", ClassID: &class2},
	}
	noonRepo.On("GetMatchByID", 9).Return(&models.NoonGameMatchWithResult{
		NoonGameMatch: &models.NoonGameMatch{ID: 9, SessionID: 3, Entries: entries},
		Entries:       entries,
	}, nil).Once()
	noonRepo.On("GetSessionByID", 3).Return(&models.NoonGameSession{ID: 3, EventID: 7}, nil).Once()
	eventRepo.On("GetEventByID", 7).Return(&models.Event{ID: 7}, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "match_id", Value: "9"}}
	c.Set("active_event_id", 7)
	c.Set("user", &models.User{ID: "00000000-0000-0000-0000-000000000001"})
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/noon-game/matches/9/result", bytes.NewBufferString(
		`{"rankings":[{"entry_id":21,"rank":1,"points":10},{"entry_id":21,"rank":2,"points":5}]}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")

	h.RecordMatchResult(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "重複")
	noonRepo.AssertNotCalled(t, "SaveMatchResult", mock.Anything, mock.Anything)
	noonRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestRecordMatchResultRejectsArchivedEventMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noonRepo := new(MockNoonGameRepository)
	h := handler.NewNoonGameHandler(noonRepo, new(MockClassRepository), new(MockEventRepository))

	noonRepo.On("GetMatchByID", 9).Return(&models.NoonGameMatchWithResult{
		NoonGameMatch: &models.NoonGameMatch{ID: 9, SessionID: 3},
	}, nil).Once()
	noonRepo.On("GetSessionByID", 3).Return(&models.NoonGameSession{ID: 3, EventID: 7}, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "match_id", Value: "9"}}
	c.Set("active_event_id", 8)
	c.Set("user", &models.User{ID: "00000000-0000-0000-0000-000000000001"})
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/noon-game/matches/9/result", bytes.NewBufferString(`{"winner":"home"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.RecordMatchResult(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	noonRepo.AssertNotCalled(t, "SaveMatchResult", mock.Anything, mock.Anything)
	noonRepo.AssertExpectations(t)
}

func TestSaveMatchKeepsParticipantIDsAndReportsLockedParticipants(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noonRepo := new(MockNoonGameRepository)
	h := handler.NewNoonGameHandler(noonRepo, new(MockClassRepository), new(MockEventRepository))

	noonRepo.On("GetSessionByID", 3).Return(&models.NoonGameSession{ID: 3, EventID: 7}, nil).Once()
	noonRepo.On("SaveMatch", mock.MatchedBy(func(match *models.NoonGameMatch) bool {
		return len(match.Entries) == 2 && match.Entries[0].ID == 21 && match.Entries[1].ID == 22
	})).Return(nil, repository.ErrNoonGameMatchParticipantsLocked).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "session_id", Value: "3"}, {Key: "match_id", Value: "9"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/root/events/7/noon-game/sessions/3/matches/9", bytes.NewBufferString(
		`{"status":"completed","home_side":{"type":"class","class_id":101},"away_side":{"type":"class","class_id":102},"participants":[{"id":21,"type":"class","class_id":101},{"id":22,"type":"class","class_id":102}]}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")

	h.SaveMatch(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "参加者")
	noonRepo.AssertExpectations(t)
}
