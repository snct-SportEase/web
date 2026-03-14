package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	classRepo repository.ClassRepository
	eventRepo repository.EventRepository
	sportRepo repository.SportRepository
	tournRepo repository.TournamentRepository
}

func NewStatisticsHandler(classRepo repository.ClassRepository, eventRepo repository.EventRepository, sportRepo repository.SportRepository, tournRepo repository.TournamentRepository) *StatisticsHandler {
	return &StatisticsHandler{
		classRepo: classRepo,
		eventRepo: eventRepo,
		sportRepo: sportRepo,
		tournRepo: tournRepo,
	}
}

func (h *StatisticsHandler) GetOverallAttendanceRate(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get classes"})
		return
	}

	totalAttendance := 0
	totalStudents := 0
	for _, class := range classes {
		totalAttendance += class.AttendCount
		totalStudents += class.StudentCount
	}

	rate := 0.0
	if totalStudents > 0 {
		rate = float64(totalAttendance) / float64(totalStudents) * 100
	}

	c.JSON(http.StatusOK, gin.H{"attendance_rate": rate})
}

func (h *StatisticsHandler) GetParticipationRateBySport(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	sports, err := h.sportRepo.GetAllSports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports"})
		return
	}

	result := make(map[string]float64)
	for _, sport := range sports {
		tournaments, err := h.tournRepo.GetTournamentsByEventAndSportID(eventID, sport.ID)
		if err != nil {
			continue
		}

		participantCount := 0
		for _, tourn := range tournaments {
			teams, err := h.tournRepo.GetTeamsByTournamentID(tourn.ID)
			if err != nil {
				continue
			}
			participantCount += len(teams)
		}

		classes, _ := h.classRepo.GetAllClasses(eventID)
		totalPossible := len(classes)
		rate := 0.0
		if totalPossible > 0 {
			rate = float64(participantCount) / float64(totalPossible) * 100
		}
		result[sport.Name] = rate
	}

	c.JSON(http.StatusOK, result)
}

func (h *StatisticsHandler) GetClassScoreTrends(c *gin.Context) {
	events, err := h.eventRepo.GetAllEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events"})
		return
	}

	result := make(map[string][]*models.ClassScore)
	for _, event := range events {
		scores, err := h.classRepo.GetClassScoresByEvent(event.ID)
		if err != nil {
			continue
		}
		result[event.Name] = scores
	}

	c.JSON(http.StatusOK, result)
}

func (h *StatisticsHandler) GetRealtimeEventProgress(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	tournaments, err := h.tournRepo.GetTournamentsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tournaments"})
		return
	}

	progress := make(map[string]string)
	for _, tourn := range tournaments {
		sport, err := h.sportRepo.GetSportByID(tourn.SportID)
		if err != nil {
			continue
		}
		progress[sport.Name] = "進行中" // Assuming status
	}

	c.JSON(http.StatusOK, progress)
}
