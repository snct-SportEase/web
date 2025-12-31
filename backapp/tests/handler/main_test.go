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

func (m *MockClassRepository) GetClassByRepRole(userID string, eventID int) (*models.Class, error) {
	args := m.Called(userID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Class), args.Error(1)
}

func (m *MockClassRepository) GetClassMembers(classID int) ([]*models.User, error) {
	args := m.Called(classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockClassRepository) SetNoonGamePoints(eventID int, points map[int]int) error {
	args := m.Called(eventID, points)
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

func (m *MockEventRepository) SetRainyMode(eventID int, isRainyMode bool) error {
	args := m.Called(eventID, isRainyMode)
	return args.Error(0)
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

func (m *MockWhitelistRepository) DeleteWhitelistedEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockWhitelistRepository) DeleteWhitelistedEmails(emails []string) error {
	args := m.Called(emails)
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

func (m *MockTeamRepository) GetTeamsByClassID(classID int, eventID int) ([]*models.TeamWithSport, error) {
	args := m.Called(classID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TeamWithSport), args.Error(1)
}

func (m *MockTeamRepository) GetTeamByClassAndSport(classID int, sportID int, eventID int) (*models.Team, error) {
	args := m.Called(classID, sportID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) AddTeamMember(teamID int, userID string) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamMembers(teamID int) ([]*models.User, error) {
	args := m.Called(teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockTeamRepository) RemoveTeamMember(teamID int, userID string) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) UpdateTeamCapacity(eventID int, sportID int, classID int, minCapacity *int, maxCapacity *int) error {
	args := m.Called(eventID, sportID, classID, minCapacity, maxCapacity)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamCapacity(eventID int, sportID int, classID int) (*models.Team, error) {
	args := m.Called(eventID, sportID, classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) ConfirmTeamMember(teamID int, userID string) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) GetConfirmedTeamMembers(teamID int) ([]*models.User, error) {
	args := m.Called(teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockTeamRepository) GetConfirmedTeamMembersCount(teamID int) (int, error) {
	args := m.Called(teamID)
	return args.Int(0), args.Error(1)
}

type MockTournamentRepository struct {
	mock.Mock
}

func (m *MockTournamentRepository) SaveTournament(eventID int, sportID int, sportName string, tournamentData *models.TournamentData, teams []*models.Team) error {
	args := m.Called(eventID, sportID, sportName, tournamentData, teams)
	return args.Error(0)
}

func (m *MockTournamentRepository) DeleteTournamentsByEventID(eventID int) error {
	args := m.Called(eventID)
	return args.Error(0)
}

func (m *MockTournamentRepository) DeleteTournamentsByEventAndSportID(eventID int, sportID int) error {
	args := m.Called(eventID, sportID)
	return args.Error(0)
}

func (m *MockTournamentRepository) GetTournamentsByEventID(eventID int) ([]*models.Tournament, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) GetMatchesForTeam(eventID int, teamID int) ([]*models.MatchDetail, error) {
	args := m.Called(eventID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MatchDetail), args.Error(1)
}

func (m *MockTournamentRepository) UpdateMatchStartTime(matchID int, startTime string) error {
	args := m.Called(matchID, startTime)
	return args.Error(0)
}

func (m *MockTournamentRepository) UpdateMatchRainyModeStartTime(matchID int, rainyModeStartTime string) error {
	args := m.Called(matchID, rainyModeStartTime)
	return args.Error(0)
}

func (m *MockTournamentRepository) UpdateMatchStatus(matchID int, status string) error {
	args := m.Called(matchID, status)
	return args.Error(0)
}

func (m *MockTournamentRepository) UpdateMatchResult(matchID, team1Score, team2Score, winnerID int) error {
	args := m.Called(matchID, team1Score, team2Score, winnerID)
	return args.Error(0)
}

func (m *MockTournamentRepository) UpdateMatchResultForCorrection(matchID, team1Score, team2Score, winnerID int) error {
	args := m.Called(matchID, team1Score, team2Score, winnerID)
	return args.Error(0)
}

func (m *MockTournamentRepository) GetTournamentIDByMatchID(matchID int) (int, error) {
	args := m.Called(matchID)
	return args.Int(0), args.Error(1)
}

func (m *MockTournamentRepository) ApplyRainyModeStartTimes(eventID int) error {
	args := m.Called(eventID)
	return args.Error(0)
}

func (m *MockTournamentRepository) IsMatchResultAlreadyEntered(matchID int) (bool, error) {
	args := m.Called(matchID)
	return args.Bool(0), args.Error(1)
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

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) CreateNotification(title, body, createdBy string, eventID *int) (int64, error) {
	args := m.Called(title, body, createdBy, eventID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) AddNotificationTargets(notificationID int64, roles []string) error {
	args := m.Called(notificationID, roles)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotificationsForAccess(roleNames []string, authorID string, includeAuthored bool, limit int) ([]models.Notification, error) {
	args := m.Called(roleNames, authorID, includeAuthored, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetUserIDsByRoles(roleNames []string) ([]string, error) {
	args := m.Called(roleNames)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockNotificationRepository) GetPushSubscriptionsByUserIDs(userIDs []string) ([]models.PushSubscription, error) {
	args := m.Called(userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PushSubscription), args.Error(1)
}

func (m *MockNotificationRepository) GetPushSubscriptionsByUserID(userID string) ([]models.PushSubscription, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PushSubscription), args.Error(1)
}

func (m *MockNotificationRepository) UpsertPushSubscription(userID, endpoint, authKey, p256dhKey string) error {
	args := m.Called(userID, endpoint, authKey, p256dhKey)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeletePushSubscription(userID, endpoint string) error {
	args := m.Called(userID, endpoint)
	return args.Error(0)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetAllRoles() ([]models.Role, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Role), args.Error(1)
}

type MockRainyModeRepository struct {
	mock.Mock
}

func (m *MockRainyModeRepository) GetSettingsByEventID(eventID int) ([]*models.RainyModeSetting, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RainyModeSetting), args.Error(1)
}

func (m *MockRainyModeRepository) GetSetting(eventID int, sportID int, classID int) (*models.RainyModeSetting, error) {
	args := m.Called(eventID, sportID, classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RainyModeSetting), args.Error(1)
}

func (m *MockRainyModeRepository) UpsertSetting(setting *models.RainyModeSetting) error {
	args := m.Called(setting)
	return args.Error(0)
}

func (m *MockRainyModeRepository) DeleteSetting(eventID int, sportID int, classID int) error {
	args := m.Called(eventID, sportID, classID)
	return args.Error(0)
}
