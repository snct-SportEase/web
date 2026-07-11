package middleware_test

import (
	"backapp/internal/middleware"
	"backapp/internal/models"
	"backapp/internal/repository"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestActiveEventStatusRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const activeEventQuery = "SELECT event_id FROM active_event WHERE id = 1"
	const eventQuery = "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores, duplicate_registration_threshold FROM events WHERE id = ?"
	columns := []string{"id", "name", "year", "season", "start_date", "end_date", "is_rainy_mode", "competition_guidelines_pdf_url", "survey_url", "is_survey_published", "status", "hide_scores", "duplicate_registration_threshold"}

	for _, tc := range []struct {
		name       string
		status     string
		wantStatus int
	}{
		{name: "allows active event", status: models.EventStatusActive, wantStatus: http.StatusNoContent},
		{name: "rejects preparing event", status: models.EventStatusPreparing, wantStatus: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta(activeEventQuery)).WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(1))
			mock.ExpectQuery(regexp.QuoteMeta(eventQuery)).WithArgs(1).WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "大会", 2026, "spring", nil, nil, false, nil, nil, false, tc.status, false, 31))

			router := gin.New()
			router.Use(middleware.ActiveEventStatusRequired(repository.NewEventRepository(db), models.EventStatusActive))
			router.PUT("/result", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/result", nil))
			if recorder.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, recorder.Code)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestActiveEventStatusRequiredRejectsMissingActiveEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id FROM active_event WHERE id = 1")).WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(nil))
	router := gin.New()
	router.Use(middleware.ActiveEventStatusRequired(repository.NewEventRepository(db), models.EventStatusActive))
	router.PUT("/result", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/result", nil))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestActiveEventStatusRequiredReturnsServerErrorForRepositoryFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id FROM active_event WHERE id = 1")).WillReturnError(assert.AnError)
	router := gin.New()
	router.Use(middleware.ActiveEventStatusRequired(repository.NewEventRepository(db), models.EventStatusActive))
	router.PUT("/result", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/result", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
