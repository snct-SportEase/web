package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"backapp/internal/websocket"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TournamentHandler struct {
	tournRepo  repository.TournamentRepository
	sportRepo  repository.SportRepository
	teamRepo   repository.TeamRepository
	classRepo  repository.ClassRepository
	hubManager *websocket.HubManager
}

func NewTournamentHandler(tournRepo repository.TournamentRepository, sportRepo repository.SportRepository, teamRepo repository.TeamRepository, classRepo repository.ClassRepository, hubManager *websocket.HubManager) *TournamentHandler {
	return &TournamentHandler{
		tournRepo:  tournRepo,
		sportRepo:  sportRepo,
		teamRepo:   teamRepo,
		classRepo:  classRepo,
		hubManager: hubManager,
	}
}

func (h *TournamentHandler) GetTournamentsByEventHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	tournaments, err := h.tournRepo.GetTournamentsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tournaments", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tournaments)
}

func (h *TournamentHandler) GenerateAllTournamentsPreviewHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	sports, err := h.sportRepo.GetSportsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports for the event", "details": err.Error()})
		return
	}

	generatedTournaments := make([]models.GeneratedTournament, 0)

	for _, eventSport := range sports {
		roundBusyClasses := make(map[int]map[int]bool)
		sport, err := h.sportRepo.GetSportByID(eventSport.SportID)
		if err != nil {
			continue // Log this error
		}

		teams, err := h.sportRepo.GetTeamsBySportID(sport.ID)
		if err != nil {
			continue // Log this error
		}

		tournamentData, shuffledTeams, err := generateTournamentStructure(teams, roundBusyClasses)
		if err != nil {
			continue
		}

		// Convert []*models.Team to []models.Team
		teamsSlice := make([]models.Team, len(shuffledTeams))
		for i, t := range shuffledTeams {
			if t != nil {
				teamsSlice[i] = *t
			}
		}
		generatedTournaments = append(generatedTournaments, models.GeneratedTournament{
			EventID:        eventID,
			SportID:        sport.ID,
			SportName:      sport.Name,
			TournamentData: *tournamentData,
			ShuffledTeams:  teamsSlice,
		})
	}

	c.JSON(http.StatusOK, generatedTournaments)
}

func (h *TournamentHandler) BulkCreateTournamentsHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var tournamentsToCreate []models.GeneratedTournament
	if err := c.ShouldBindJSON(&tournamentsToCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// First, delete existing tournaments for this event
	if err := h.tournRepo.DeleteTournamentsByEventID(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing tournaments", "details": err.Error()})
		return
	}

	// Then, save the new tournaments
	for _, t := range tournamentsToCreate {
		// Convert []models.Team to []*models.Team
		shuffledTeamsPtr := make([]*models.Team, len(t.ShuffledTeams))
		for i := range t.ShuffledTeams {
			shuffledTeamsPtr[i] = &t.ShuffledTeams[i]
		}
		err := h.tournRepo.SaveTournament(t.EventID, t.SportID, t.SportName, &t.TournamentData, shuffledTeamsPtr)
		if err != nil {
			// Attempt to rollback or handle partial save might be needed here
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tournament for sport " + t.SportName, "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournaments created successfully."})
}

func (h *TournamentHandler) GenerateAllTournamentsHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.tournRepo.DeleteTournamentsByEventID(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing tournaments", "details": err.Error()})
		return
	}

	sports, err := h.sportRepo.GetSportsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports for the event", "details": err.Error()})
		return
	}

	for _, eventSport := range sports {
		roundBusyClasses := make(map[int]map[int]bool)
		sport, err := h.sportRepo.GetSportByID(eventSport.SportID)
		if err != nil {
			continue
		}

		teams, err := h.sportRepo.GetTeamsBySportID(sport.ID)
		if err != nil {
			continue
		}

		tournamentData, shuffledTeams, err := generateTournamentStructure(teams, roundBusyClasses)
		if err != nil {
			continue
		}

		err = h.tournRepo.SaveTournament(eventID, sport.ID, sport.Name, tournamentData, shuffledTeams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tournament for sport " + sport.Name, "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournaments generated successfully for all eligible sports."})
}

func generateTournamentStructure(teams []*models.Team, roundBusyClasses map[int]map[int]bool) (*models.TournamentData, []*models.Team, error) {
	availableTeams := make([]*models.Team, 0)
	if roundBusyClasses[0] == nil {
		roundBusyClasses[0] = make(map[int]bool)
	}

	for _, team := range teams {
		if !roundBusyClasses[0][team.ClassID] {
			availableTeams = append(availableTeams, team)
		}
	}

	if len(availableTeams) <= 1 || (len(availableTeams)&(len(availableTeams)-1)) != 0 {
		return nil, nil, fmt.Errorf("invalid number of available teams to form a tournament: %d", len(availableTeams))
	}

	numRounds := int(math.Log2(float64(len(availableTeams))))

	for _, team := range availableTeams {
		for i := 0; i < numRounds; i++ {
			if roundBusyClasses[i] == nil {
				roundBusyClasses[i] = make(map[int]bool)
			}
			roundBusyClasses[i][team.ClassID] = true
		}
	}

	shuffledTeams := make([]*models.Team, len(availableTeams))
	copy(shuffledTeams, availableTeams)

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	r.Shuffle(len(shuffledTeams), func(i, j int) {
		shuffledTeams[i], shuffledTeams[j] = shuffledTeams[j], shuffledTeams[i]
	})

	contestants := make(map[string]models.Contestant)
	for i, team := range shuffledTeams {
		contestantID := "c" + strconv.Itoa(i)
		contestants[contestantID] = models.Contestant{
			Players: []models.Player{{Title: team.Name}},
		}
	}

	rounds := make([]models.Round, numRounds)

	var roundNames []string
	switch numRounds {
	case 1:
		roundNames = []string{"決勝"}
	case 2:
		roundNames = []string{"準決勝", "決勝"}
	case 3:
		roundNames = []string{"一回戦", "準決勝", "決勝"}
	case 4:
		roundNames = []string{"一回戦", "二回戦", "準決勝", "決勝"}
	default:
		roundNames = make([]string, numRounds)
		for i := 0; i < numRounds-2; i++ {
			roundNames[i] = strconv.Itoa(i+1) + "回戦"
		}
		if numRounds > 1 {
			roundNames[numRounds-2] = "準決勝"
		}
		if numRounds > 0 {
			roundNames[numRounds-1] = "決勝"
		}
	}

	for i := 0; i < numRounds; i++ {
		rounds[i] = models.Round{Name: roundNames[i]}
	}

	matches := []models.Match{}

	matchOrder := 0
	for i := 0; i < len(shuffledTeams); i += 2 {
		contestant1ID := "c" + strconv.Itoa(i)
		contestant2ID := "c" + strconv.Itoa(i+1)
		matches = append(matches, models.Match{
			RoundIndex: 0,
			Order:      matchOrder,
			Sides: []models.Side{
				{ContestantID: contestant1ID},
				{ContestantID: contestant2ID},
			},
		})
		matchOrder++
	}

	numMatchesInRound := len(shuffledTeams) / 2
	for roundIndex := 1; roundIndex < numRounds; roundIndex++ {
		matchOrder = 0
		numMatchesInRound /= 2
		for i := 0; i < numMatchesInRound; i++ {
			matches = append(matches, models.Match{
				RoundIndex: roundIndex,
				Order:      matchOrder,
				Sides:      []models.Side{{}, {}},
			})
			matchOrder++
		}
	}

	if numRounds >= 2 {
		matches = append(matches, models.Match{
			RoundIndex:    numRounds - 1,
			Order:         1,
			IsBronzeMatch: true,
			Sides:         []models.Side{{}, {}},
		})
	}

	tournamentData := models.TournamentData{
		Rounds:      rounds,
		Matches:     matches,
		Contestants: contestants,
	}

	return &tournamentData, shuffledTeams, nil
}

