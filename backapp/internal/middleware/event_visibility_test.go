package middleware

import (
	"backapp/internal/models"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubEventVisibilityReader struct {
	activeID int
	event    *models.Event
	err      error
}

func (s stubEventVisibilityReader) GetActiveEvent() (int, error) {
	return s.activeID, s.err
}

func (s stubEventVisibilityReader) GetEventByID(int) (*models.Event, error) {
	return s.event, s.err
}

func runEventVisibility(t *testing.T, path string, user *models.User, reader EventVisibilityReader) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user", user) })
	router.GET("/events/:event_id/resources", EventVisibilityRequired(reader), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestEventVisibilityRequiredHidesUnannouncedEventFromStudent(t *testing.T) {
	student := &models.User{Roles: []models.Role{{Name: "student"}}}
	for _, status := range []string{models.EventStatusPreparing, models.EventStatusTesting} {
		recorder := runEventVisibility(t, "/events/9/resources", student, stubEventVisibilityReader{
			event: &models.Event{ID: 9, Status: status},
		})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("status %q: got %d, want %d", status, recorder.Code, http.StatusNotFound)
		}
	}
}

func TestEventVisibilityRequiredAllowsAnnouncedEventForStudent(t *testing.T) {
	recorder := runEventVisibility(t, "/events/9/resources", &models.User{Roles: []models.Role{{Name: "student"}}}, stubEventVisibilityReader{
		event: &models.Event{ID: 9, Status: models.EventStatusUpcoming},
	})
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("got %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestEventVisibilityRequiredAllowsAdminWithoutLookup(t *testing.T) {
	recorder := runEventVisibility(t, "/events/9/resources", &models.User{Roles: []models.Role{{Name: "admin"}}}, stubEventVisibilityReader{
		err: errors.New("must not be called"),
	})
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("got %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
