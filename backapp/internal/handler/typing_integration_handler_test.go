package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backapp/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type typingIntegrationRepositoryStub struct {
	snapshot *models.TypingCompetitionSnapshot
	err      error
}

func (s *typingIntegrationRepositoryStub) GetActiveEvent() (*models.TypingEvent, []*models.TypingSport, error) {
	return nil, nil, s.err
}

func (s *typingIntegrationRepositoryStub) GetCompetitionSnapshot(_, _ int) (*models.TypingCompetitionSnapshot, error) {
	return s.snapshot, s.err
}

func (s *typingIntegrationRepositoryStub) SetTeamEntryOrder(_ int, _ []string) error {
	return s.err
}

func TestTypingSnapshotWarningsReportsIncompleteRoster(t *testing.T) {
	snapshot := &models.TypingCompetitionSnapshot{
		Teams: []*models.TypingTeam{
			{ID: 10, Name: "1-A", Players: []*models.TypingPlayer{{ID: "a"}}},
		},
		Tournaments: []*models.TypingTournament{},
	}

	warnings := typingSnapshotWarnings(snapshot)
	assert.Contains(t, warnings, "expected 6 teams, got 1")
	assert.Contains(t, warnings, "team 10 (1-A) must have 3 confirmed players, got 1")
	assert.Contains(t, warnings, "no matches are registered")
}

func TestGetCompetitionSnapshotRejectsInvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTypingIntegrationHandler(&typingIntegrationRepositoryStub{err: errors.New("must not be called")})
	router := gin.New()
	router.GET("/events/:event_id/sports/:sport_id", handler.GetCompetitionSnapshot)

	request := httptest.NewRequest(http.MethodGet, "/events/0/sports/not-a-number", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetCompetitionSnapshotReturnsVersionedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	snapshot := &models.TypingCompetitionSnapshot{
		Event:       &models.TypingEvent{ID: 1},
		Sport:       &models.TypingSport{ID: 2},
		Teams:       make([]*models.TypingTeam, 6),
		Tournaments: []*models.TypingTournament{{Matches: []*models.TypingMatch{{ID: 3}}}},
	}
	for i := range snapshot.Teams {
		snapshot.Teams[i] = &models.TypingTeam{
			ID: i + 1,
			Players: []*models.TypingPlayer{
				{ID: "one", EntryOrder: 1},
				{ID: "two", EntryOrder: 2},
				{ID: "three", EntryOrder: 3},
			},
		}
	}

	handler := NewTypingIntegrationHandler(&typingIntegrationRepositoryStub{snapshot: snapshot})
	router := gin.New()
	router.GET("/events/:event_id/sports/:sport_id", handler.GetCompetitionSnapshot)

	request := httptest.NewRequest(http.MethodGet, "/events/1/sports/2", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"api_version":"v1"`)
	assert.Contains(t, recorder.Body.String(), `"warnings":[]`)
}
