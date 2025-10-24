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

// --- Mocks ---

type MockClassRepository struct {
	mock.Mock
}

func (m *MockClassRepository) GetAllClasses(eventID int) ([]*models.Class, error) {
	args := m.Called(eventID)
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

func (m *MockClassRepository) CreateClasses(eventID int, classNames []string) error {
	args := m.Called(eventID, classNames)
	return args.Error(0)
}

func (m *MockClassRepository) UpdateStudentCounts(eventID int, counts map[int]int) error {
    args := m.Called(eventID, counts)
    return args.Error(0)
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
