package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestClassProgressSummaryRejectsInvalidViewerBeforeLoadingData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	classID, eventID, foreignEventID := 10, 1, 2
	for _, tc := range []struct {
		name       string
		user       any
		setUser    bool
		activeID   int
		activeErr  error
		class      *models.Class
		classErr   error
		wantStatus int
	}{
		{name: "missing session", wantStatus: 401},
		{name: "invalid viewer type", user: "student", setUser: true, wantStatus: 500},
		{name: "nil viewer", user: (*models.User)(nil), setUser: true, wantStatus: 500},
		{name: "no class", user: &models.User{ID: "student"}, setUser: true, activeID: eventID, wantStatus: 403},
		{name: "no active event", user: &models.User{ID: "student", ClassID: &classID}, setUser: true, wantStatus: 404},
		{name: "active event error", user: &models.User{ID: "student", ClassID: &classID}, setUser: true, activeErr: errors.New("unavailable"), wantStatus: 500},
		{name: "missing class", user: &models.User{ID: "student", ClassID: &classID}, setUser: true, activeID: eventID, wantStatus: 403},
		{name: "class error", user: &models.User{ID: "student", ClassID: &classID}, setUser: true, activeID: eventID, classErr: errors.New("unavailable"), wantStatus: 500},
		{name: "class from another event", user: &models.User{ID: "student", ClassID: &classID}, setUser: true, activeID: eventID, class: &models.Class{ID: classID, EventID: &foreignEventID}, wantStatus: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			classes, events := new(MockClassRepository), new(MockEventRepository)
			teams, tournaments := new(MockTeamRepository), new(MockTournamentRepository)
			if user, ok := tc.user.(*models.User); tc.setUser && ok && user != nil {
				events.On("GetActiveEvent").Return(tc.activeID, tc.activeErr).Once()
				if tc.activeErr == nil && tc.activeID != 0 && user.ClassID != nil {
					classes.On("GetClassByID", classID).Return(tc.class, tc.classErr).Once()
				}
			}
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodGet, "/api/student/class-progress?view=summary&class_id=999&event_id=999", nil)
			if tc.setUser {
				context.Set("user", tc.user)
			}
			handler.NewClassHandler(classes, events, teams, tournaments).GetClassProgress(context)
			require.Equal(t, tc.wantStatus, response.Code)
			require.NotContains(t, response.Body.String(), `"progress"`)
			require.NotContains(t, response.Body.String(), `"student_count"`)
			// No count, member, assignment, or match query may precede validation.
			require.Len(t, teams.Calls, 0)
			require.Len(t, tournaments.Calls, 0)
			for _, call := range classes.Calls {
				require.Equal(t, "GetClassByID", call.Method)
			}
			classes.AssertExpectations(t)
			events.AssertExpectations(t)
		})
	}
}
