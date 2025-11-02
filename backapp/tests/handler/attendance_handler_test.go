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
)

// --- Tests ---

func TestGetClassDetailsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		expectedDetails := &models.ClassDetails{
			ID:               1,
			Name:             "Test Class",
			StudentCount:     25,
			AttendancePoints: 10,
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassDetails", 1, 1).Return(expectedDetails, nil).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "classID", Value: "1"}}

		h.GetClassDetailsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		expectedJSON, _ := json.Marshal(expectedDetails)
		assert.JSONEq(t, string(expectedJSON), w.Body.String())
		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	// ... other tests for GetClassDetailsHandler are unchanged
}

func TestRegisterAttendanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		activeEventID := 1

		reqBody := handler.RegisterAttendanceRequest{
			ClassID:         1,
			AttendanceCount: 20,
		}
		class := &models.Class{ID: 1, EventID: &activeEventID, Name: "Test Class", StudentCount: 25}
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", reqBody.ClassID).Return(class, nil).Once()
		mockClassRepo.On("UpdateAttendance", reqBody.ClassID, activeEventID, reqBody.AttendanceCount).Return(10, nil).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.RegisterAttendanceHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully registered attendance")
		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Class does not belong to active event", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		activeEventID := 1
		differentEventID := 2

		reqBody := handler.RegisterAttendanceRequest{
			ClassID:         1,
			AttendanceCount: 20,
		}
		class := &models.Class{ID: 1, EventID: &differentEventID, Name: "Test Class", StudentCount: 25}

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", reqBody.ClassID).Return(class, nil).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.RegisterAttendanceHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Class does not belong to the active event")
		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Attendance Exceeds Student Count", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		activeEventID := 1

		reqBody := handler.RegisterAttendanceRequest{
			ClassID:         1,
			AttendanceCount: 30,
		}
		class := &models.Class{ID: 1, EventID: &activeEventID, Name: "Test Class", StudentCount: 25}

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", reqBody.ClassID).Return(class, nil).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.RegisterAttendanceHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "cannot exceed student count")
		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	// ... other tests for RegisterAttendanceHandler should also be updated similarly
}
