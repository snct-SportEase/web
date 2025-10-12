package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TournamentHandler struct {
	tournRepo repository.TournamentRepository
	sportRepo repository.SportRepository
	teamRepo  repository.TeamRepository
	classRepo repository.ClassRepository
}

func NewTournamentHandler(tournRepo repository.TournamentRepository, sportRepo repository.SportRepository, teamRepo repository.TeamRepository, classRepo repository.ClassRepository) *TournamentHandler {
	return &TournamentHandler{
		tournRepo: tournRepo,
		sportRepo: sportRepo,
		teamRepo:  teamRepo,
		classRepo: classRepo,
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
		sport, err := h.sportRepo.GetSportByID(eventSport.SportID)
		if err != nil {
			continue
		}

		teams, err := h.sportRepo.GetTeamsBySportID(sport.ID)
		if err != nil {
			continue
		}

		if len(teams) <= 1 {
			continue
		}

		if (len(teams) & (len(teams) - 1)) != 0 {
			continue
		}

		tournamentData, shuffledTeams := generateTournamentStructure(teams)

		err = h.tournRepo.SaveTournament(eventID, sport.ID, sport.Name, tournamentData, shuffledTeams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tournament for sport " + sport.Name, "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournaments generated successfully for all eligible sports."})
}

func generateTournamentStructure(teams []*models.Team) (*models.TournamentData, []*models.Team) {
	shuffledTeams := make([]*models.Team, len(teams))
	copy(shuffledTeams, teams)

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

	numRounds := int(math.Log2(float64(len(shuffledTeams))))
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

	return &tournamentData, shuffledTeams
}
