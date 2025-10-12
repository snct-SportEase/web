package handler

import (
	"backapp/internal/models"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateTournamentHandler handles the request to generate a tournament.
func GenerateTournamentHandler(c *gin.Context) {
	var teams []models.TeamName
	if err := c.ShouldBindJSON(&teams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(teams) <= 1 || (len(teams)&(len(teams)-1)) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number of teams must be a power of two and greater than 1"})
		return
	}

	// Shuffle teams
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	r.Shuffle(len(teams), func(i, j int) {
		teams[i], teams[j] = teams[j], teams[i]
	})

	contestants := make(map[string]models.Contestant)
	for i, team := range teams {
		contestantID := "c" + strconv.Itoa(i)
		className := strings.Split(team.Name, "_")[0]
		contestants[contestantID] = models.Contestant{
			Players: []models.Player{{Title: className}},
		}
	}

	numRounds := int(math.Log2(float64(len(teams))))
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
		// Generic naming for other cases
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

	// First round
	matchOrder := 0
	for i := 0; i < len(teams); i += 2 {
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

	// Subsequent rounds
	numMatchesInRound := len(teams) / 2
	for roundIndex := 1; roundIndex < numRounds; roundIndex++ {
		matchOrder = 0
		numMatchesInRound /= 2
		for i := 0; i < numMatchesInRound; i++ {
			matches = append(matches, models.Match{
				RoundIndex: roundIndex,
				Order:      matchOrder,
				Sides:      []models.Side{{}, {}}, // Empty sides for subsequent rounds
			})
			matchOrder++
		}
	}

	// Add bronze match if there are enough rounds for it (a semi-final exists)
	if numRounds >= 2 {
		matches = append(matches, models.Match{
			RoundIndex:    numRounds - 1, // Same round as the final
			Order:         1,             // After the final
			IsBronzeMatch: true,
			Sides:         []models.Side{{}, {}},
		})
	}

	tournamentData := models.TournamentData{
		Rounds:      rounds,
		Matches:     matches,
		Contestants: contestants,
	}

	c.JSON(http.StatusOK, tournamentData)
}
