package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/repository"
	"backapp/internal/websocket"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTournamentSeedHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	newHandler := func() (*handler.TournamentHandler, *MockTournamentRepository) {
		repo := new(MockTournamentRepository)
		return handler.NewTournamentHandler(repo, new(MockSportRepository), new(MockTeamRepository), new(MockClassRepository), new(MockEventRepository), websocket.NewHubManager()), repo
	}
	newContext := func(method, body string) (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "2"}, {Key: "tournament_id", Value: "22"}}
		c.Request = httptest.NewRequest(method, "/api/root/events/2/tournaments/22/seeds", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		return c, w
	}

	t.Run("returns persisted seed order", func(t *testing.T) {
		h, repo := newHandler()
		repo.On("GetTournamentSeedOrder", 2, 22).Return([]repository.TournamentSeedTeam{{TeamID: 1, TeamName: "A"}, {TeamID: 2, TeamName: "B"}}, nil).Once()
		c, w := newContext(http.MethodGet, "")
		h.GetTournamentSeedsHandler(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"team_id":1`)
		repo.AssertExpectations(t)
	})
	t.Run("saves reordered teams", func(t *testing.T) {
		h, repo := newHandler()
		repo.On("UpdateTournamentSeedOrder", 2, 22, []int{1, 2}, []int{2, 1}).Return(nil).Once()
		c, w := newContext(http.MethodPut, `{"expected_team_ids":[1,2],"team_ids":[2,1]}`)
		h.UpdateTournamentSeedsHandler(c)
		require.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})
	t.Run("rejects started matches", func(t *testing.T) {
		h, repo := newHandler()
		repo.On("UpdateTournamentSeedOrder", 2, 22, []int{1, 2}, []int{2, 1}).Return(repository.ErrTournamentSeedLocked).Once()
		c, w := newContext(http.MethodPut, `{"expected_team_ids":[1,2],"team_ids":[2,1]}`)
		h.UpdateTournamentSeedsHandler(c)
		require.Equal(t, http.StatusConflict, w.Code)
		repo.AssertExpectations(t)
	})
}
