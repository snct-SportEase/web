package middleware

import (
	"backapp/internal/models"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubTestRunStatusReader struct {
	status *models.EventTestRunStatus
	err    error
}

func (s stubTestRunStatusReader) GetStatus(context.Context) (*models.EventTestRunStatus, error) {
	return s.status, s.err
}

func runTestRunIsolation(t *testing.T, user *models.User, reader TestRunStatusReader) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
	}, TestRunIsolation(reader))
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/protected", nil))
	return recorder
}

func TestTestRunIsolationBlocksStudentsDuringTestRun(t *testing.T) {
	recorder := runTestRunIsolation(t, &models.User{Roles: []models.Role{{Name: "student"}}}, stubTestRunStatusReader{
		status: &models.EventTestRunStatus{State: models.EventTestRunStateTesting},
	})
	if recorder.Code != http.StatusLocked {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusLocked)
	}
}

func TestTestRunIsolationAllowsAdminsDuringTestRun(t *testing.T) {
	recorder := runTestRunIsolation(t, &models.User{Roles: []models.Role{{Name: "admin"}}}, stubTestRunStatusReader{
		status: &models.EventTestRunStatus{State: models.EventTestRunStateTesting},
	})
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestTestRunIsolationAllowsStudentsAfterRestoreWhileNotificationsPaused(t *testing.T) {
	recorder := runTestRunIsolation(t, &models.User{Roles: []models.Role{{Name: "student"}}}, stubTestRunStatusReader{
		status: &models.EventTestRunStatus{State: models.EventTestRunStateAwaitingNotificationResume},
	})
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestTestRunIsolationFailsClosedForStudents(t *testing.T) {
	recorder := runTestRunIsolation(t, &models.User{Roles: []models.Role{{Name: "student"}}}, stubTestRunStatusReader{
		err: errors.New("database unavailable"),
	})
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
