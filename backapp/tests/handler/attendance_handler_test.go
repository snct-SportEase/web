package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockClassRepository struct {
	mock.Mock
}

func (m *MockClassRepository) GetAllClasses() ([]*models.Class, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Class), args.Error(1)
}

func (m *MockClassRepository) GetClassDetails(classID, eventID int) (*models.ClassDetails, error) {
	args := m.Called(classID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClassDetails), args.Error(1)
}

func (m *MockClassRepository) GetClassByID(classID int) (*models.Class, error) {
	args := m.Called(classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Class), args.Error(1)
}

func (m *MockClassRepository) UpdateAttendance(classID, eventID, attendanceCount int) (int, error) {
	args := m.Called(classID, eventID, attendanceCount)
	return args.Int(0), args.Error(1)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) CreateEvent(event *models.Event) (int64, error) {
	args := m.Called(event)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepository) GetAllEvents() ([]*models.Event, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Event), args.Error(1)
}

func (m *MockEventRepository) UpdateEvent(event *models.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepository) SetActiveEvent(event_id *int) error {
	args := m.Called(event_id)
	return args.Error(0)
}

func (m *MockEventRepository) GetActiveEvent() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

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

	t.Run("Invalid Class ID", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "classID", Value: "abc"}}

		h.GetClassDetailsHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid class ID format")
	})

	t.Run("No Active Event", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockEventRepo.On("GetActiveEvent").Return(0, nil).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "classID", Value: "1"}}

		h.GetClassDetailsHandler(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "No active event found")
		mockEventRepo.AssertExpectations(t)
	})
}

func TestRegisterAttendanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		reqBody := handler.RegisterAttendanceRequest{
			ClassID:         1,
			AttendanceCount: 20,
		}
		class := &models.Class{ID: 1, Name: "Test Class", StudentCount: 25}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByID", reqBody.ClassID).Return(class, nil).Once()
		mockClassRepo.On("UpdateAttendance", reqBody.ClassID, 1, reqBody.AttendanceCount).Return(10, nil).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.RegisterAttendanceHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully registered attendance")
		assert.Contains(t, w.Body.String(), "Points awarded: 10")
		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{invalid"))
		c.Request.Header.Set("Content-Type", "application/json")

		h.RegisterAttendanceHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("Attendance Exceeds Student Count", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		reqBody := handler.RegisterAttendanceRequest{
			ClassID:         1,
			AttendanceCount: 30,
		}
		class := &models.Class{ID: 1, Name: "Test Class", StudentCount: 25}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
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

	t.Run("DB Error on UpdateAttendance", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		reqBody := handler.RegisterAttendanceRequest{
			ClassID:         1,
			AttendanceCount: 20,
		}
		class := &models.Class{ID: 1, Name: "Test Class", StudentCount: 25}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByID", reqBody.ClassID).Return(class, nil).Once()
		mockClassRepo.On("UpdateAttendance", reqBody.ClassID, 1, reqBody.AttendanceCount).Return(0, errors.New("db error")).Once()

		h := handler.NewAttendanceHandler(mockClassRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.RegisterAttendanceHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to register attendance")
		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})
}
