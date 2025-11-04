package handler_test

import (
	"backapp/internal/models"
	"backapp/internal/repository"

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

func (m *MockEventRepository) GetEventByYearAndSeason(year int, season string) (*models.Event, error) {
	args := m.Called(year, season)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventRepository) CopyClassScores(fromEventID int, toEventID int) error {
	args := m.Called(fromEventID, toEventID)
	return args.Error(0)
}

func (m *MockEventRepository) GetEventByID(id int) (*models.Event, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

type MockWhitelistRepository struct {
	mock.Mock
}

func (m *MockWhitelistRepository) IsEmailWhitelisted(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockWhitelistRepository) AddWhitelistedEmail(email, role string, eventID *int) error {
	args := m.Called(email, role, eventID)
	return args.Error(0)
}

func (m *MockWhitelistRepository) GetAllWhitelistedEmails() ([]repository.WhitelistEntry, error) {
	args := m.Called()
	return args.Get(0).([]repository.WhitelistEntry), args.Error(1)
}

func (m *MockWhitelistRepository) AddWhitelistedEmails(entries []repository.WhitelistEntry) error {
	args := m.Called(entries)
	return args.Error(0)
}

func (m *MockWhitelistRepository) UpdateNullEventIDs(eventID int) error {
	args := m.Called(eventID)
	return args.Error(0)
}
