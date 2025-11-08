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

func (m *MockClassRepository) GetClassScoresByEvent(eventID int) ([]*models.ClassScore, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ClassScore), args.Error(1)
}

func (m *MockClassRepository) UpdateClassRanks(eventID int) error {
	args := m.Called(eventID)
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

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) CreateTeam(team *models.Team) (int64, error) {
	args := m.Called(team)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTeamRepository) DeleteTeamsByEventAndSportID(eventID int, sportID int) error {
	args := m.Called(eventID, sportID)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamsByUserID(userID string) ([]*models.TeamWithSport, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TeamWithSport), args.Error(1)
}

type MockSportRepository struct {
	mock.Mock
}

func (m *MockSportRepository) GetAllSports() ([]*models.Sport, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Sport), args.Error(1)
}

func (m *MockSportRepository) GetSportByID(sportID int) (*models.Sport, error) {
	args := m.Called(sportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Sport), args.Error(1)
}

func (m *MockSportRepository) CreateSport(sport *models.Sport) (int64, error) {
	args := m.Called(sport)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSportRepository) GetSportsByEventID(eventID int) ([]*models.EventSport, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EventSport), args.Error(1)
}

func (m *MockSportRepository) AssignSportToEvent(eventSport *models.EventSport) error {
	args := m.Called(eventSport)
	return args.Error(0)
}

func (m *MockSportRepository) DeleteSportFromEvent(eventID int, sportID int) error {
	args := m.Called(eventID, sportID)
	return args.Error(0)
}

func (m *MockSportRepository) GetTeamsBySportID(sportID int) ([]*models.Team, error) {
	args := m.Called(sportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Team), args.Error(1)
}

func (m *MockSportRepository) GetSportDetails(eventID int, sportID int) (*models.EventSport, error) {
	args := m.Called(eventID, sportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EventSport), args.Error(1)
}

func (m *MockSportRepository) UpdateSportDetails(eventID int, sportID int, details models.EventSport) error {
	args := m.Called(eventID, sportID, details)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUsers(query string, searchType string) ([]*models.User, error) {
	args := m.Called(query, searchType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserDisplayName(userID string, displayName string) error {
	args := m.Called(userID, displayName)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserWithRoles(userID string) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) IsEmailWhitelisted(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetRoleByEmail(email string) (string, error) {
	args := m.Called(email)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) AddUserRoleIfNotExists(userID string, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserRole(userID string, roleName string, eventID *int) error {
	args := m.Called(userID, roleName, eventID)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUserRole(userID string, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}
